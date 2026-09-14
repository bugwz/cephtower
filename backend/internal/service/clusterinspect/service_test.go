package clusterinspect

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/store"
)

type inspectExecutor struct {
	output string
	specs  []executor.CommandSpec
	fail   bool
}

func (e *inspectExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if e.fail {
		return executor.CommandResult{}, errors.New("unavailable")
	}
	return executor.CommandResult{Stdout: []byte(e.output)}, nil
}
func testInspection(t *testing.T) (*Service, *inspectExecutor, uint64) {
	t.Helper()
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "inspection.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	secret, err := security.Encrypt([]byte("fixture"), key)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	row := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: secret, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &row); err != nil {
		t.Fatal(err)
	}
	runner := &inspectExecutor{}
	return New(clusterservice.New(func() *store.Database { return db }, key, nil), runner), runner, row.ID
}
func TestLogsUseCephLogLastAndNewestFirst(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.output = `[{"name":"mon.a","rank":"0","stamp":"2026-09-14 01:00:00","seq":1,"channel":"audit","priority":"[INF]","message":"old"},{"name":"mon.b","rank":"1","stamp":"2026-09-14 01:00:01","seq":9007199254740993,"channel":"audit","priority":"[WRN]","message":"token=secret-value"}]`
	logs, err := service.Logs(context.Background(), id, "audit", "debug", 30)
	if err != nil || len(logs.Items) != 2 {
		t.Fatalf("logs=%+v err=%v", logs, err)
	}
	if logs.Items[0].Name != "mon.b" || logs.Items[0].Seq.String() != "9007199254740993" || logs.Items[0].Message != "token=[REDACTED]" {
		t.Fatalf("log=%+v", logs.Items[0])
	}
	if !reflect.DeepEqual(runner.specs[0].Args, []string{"log", "last", "30", "debug", "audit", "--format", "json"}) || runner.specs[0].Mutating {
		t.Fatalf("spec=%+v", runner.specs[0])
	}
	runner.output = "[]"
	if logs, err = service.Logs(context.Background(), id, "", "", 0); err != nil || logs.Items == nil {
		t.Fatalf("empty log failed: %v", err)
	}
}
func TestInspectionValidationAndFailures(t *testing.T) {
	service, runner, id := testInspection(t)
	for _, args := range []struct {
		channel, level string
		limit          int
	}{{"--help", "info", 10}, {"cluster", "fatal", 10}, {"audit", "debug", 501}, {"cluster", "info", -1}} {
		if _, err := service.Logs(context.Background(), id, args.channel, args.level, args.limit); err == nil {
			t.Fatal("invalid parameters accepted")
		}
	}
	if _, err := service.ConfigurationOption(context.Background(), id, "--help"); err == nil {
		t.Fatal("invalid option accepted")
	}
	if len(runner.specs) != 0 {
		t.Fatal("invalid request executed")
	}
	for _, output := range []string{"null", "not json", "{}"} {
		runner.output = output
		if _, err := service.Logs(context.Background(), id, "audit", "info", 10); err == nil {
			t.Fatalf("accepted malformed output %s", output)
		}
	}
	runner.fail = true
	if _, err := service.Logs(context.Background(), id, "cluster", "info", 10); err == nil {
		t.Fatal("failure became empty log")
	}
}
func TestConfigurationMetadataUsesHelp(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.output = `{"name":"osd_pool_default_size","type":"uint","level":"advanced","default":3,"can_update_at_runtime":true,"min":1,"max":10,"desc":"replicas","tags":["rados"]}`
	option, err := service.ConfigurationOption(context.Background(), id, "osd_pool_default_size")
	if err != nil || option["can_update_at_runtime"] != true || option["desc"] != "replicas" {
		t.Fatalf("option=%v err=%v", option, err)
	}
	if !reflect.DeepEqual(runner.specs[0].Args, []string{"config", "help", "osd_pool_default_size", "--format", "json"}) {
		t.Fatal("wrong help command")
	}
	runner.output = `{"name":"other"}`
	if _, err := service.ConfigurationOption(context.Background(), id, "osd_pool_default_size"); err == nil {
		t.Fatal("wrong option accepted")
	}
}

func TestOSDInspectionCommandsAndFailures(t *testing.T) {
	service, runner, id := testInspection(t)
	for _, tc := range []struct {
		section string
		args    []string
	}{
		{"metadata", []string{"osd", "metadata", "0", "--format", "json"}},
		{"histogram", []string{"tell", "osd.0", "perf", "histogram", "dump", "--format", "json"}},
	} {
		runner.output = `{"osd":{"latency":{"axes":[],"values":[[1,2]]}}}`
		result, err := service.OSDInspection(context.Background(), id, "0", tc.section)
		if err != nil || result["osd"] == nil {
			t.Fatalf("result=%v err=%v", result, err)
		}
		spec := runner.specs[len(runner.specs)-1]
		if spec.Mutating || !reflect.DeepEqual(spec.Args, tc.args) {
			t.Fatalf("spec=%+v", spec)
		}
	}
	count := len(runner.specs)
	for _, osd := range []string{"-1", "--help", "0;id", "1.0", ""} {
		if _, err := service.OSDInspection(context.Background(), id, osd, "metadata"); err == nil {
			t.Fatalf("accepted %q", osd)
		}
	}
	if _, err := service.OSDInspection(context.Background(), id, "0", "injectargs"); err == nil {
		t.Fatal("invalid section accepted")
	}
	if len(runner.specs) != count {
		t.Fatal("invalid command executed")
	}
	runner.output = "null"
	if _, err := service.OSDInspection(context.Background(), id, "0", "metadata"); err == nil {
		t.Fatal("null accepted")
	}
	runner.fail = true
	if _, err := service.OSDInspection(context.Background(), id, "0", "histogram"); err == nil {
		t.Fatal("offline OSD reported success")
	}
}

func TestSnapshotScheduleStatusScopeAndFailures(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.output = `[{"path":"/volumes/team/project","schedule":"1h","start":"2026-09-14T00:00:00","active":true,"created_count":4}]`
	rows, err := service.SnapshotSchedules(context.Background(), id, "data", "/", "project", "team")
	if err != nil || len(rows) != 1 || rows[0]["active"] != true {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	if spec := runner.specs[0]; spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"fs", "snap-schedule", "status", "/", "--fs=data", "--format=json", "--subvol=project", "--group=team"}) {
		t.Fatalf("spec=%+v", spec)
	}
	for _, output := range []string{"null", "{}", "not json", "[null]", `[{"path":"/","schedule":"1h","active":true}]`, `[{"path":"/","schedule":"1h","start":"2026-09-14T00:00:00","active":"false"}]`} {
		runner.output = output
		if _, err := service.SnapshotSchedules(context.Background(), id, "data", "/", "", ""); err == nil {
			t.Fatalf("accepted %q", output)
		}
	}
	runner.output = "[]"
	if rows, err := service.SnapshotSchedules(context.Background(), id, "data", "/", "", ""); err != nil || rows == nil || len(rows) != 0 {
		t.Fatalf("empty schedules rejected: %v", err)
	}
	count := len(runner.specs)
	if _, err := service.SnapshotSchedules(context.Background(), id, "data", "/", "", "team"); err == nil {
		t.Fatal("group without subvolume accepted")
	}
	if _, err := service.SnapshotSchedules(context.Background(), id, "data", "relative", "", ""); err == nil {
		t.Fatal("relative path accepted")
	}
	if len(runner.specs) != count {
		t.Fatal("invalid query executed")
	}
	runner.fail = true
	if _, err := service.SnapshotSchedules(context.Background(), id, "data", "/", "", ""); err == nil {
		t.Fatal("failure became empty schedule list")
	}
}
