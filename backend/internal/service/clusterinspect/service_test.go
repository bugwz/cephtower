package clusterinspect

import (
	"context"
	"encoding/json"
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
	run    func(executor.CommandSpec) (executor.CommandResult, error)
}

func (e *inspectExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if e.run != nil {
		return e.run(spec)
	}
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
	for _, output := range []string{"null", "not json", "{}", "[] {}", "[] failed", "[null]", "[{}]", `[{"stamp":"2026-10-03T00:00:00Z"}]`, `[{"seq":0}]`, `[{"seq":-1,"stamp":"2026-10-03T00:00:00Z"}]`} {
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

func TestInspectionReaderRejectsTrailingJSONWithoutLosingPrecision(t *testing.T) {
	s, runner, id := testInspection(t)
	for _, output := range []string{`{"value":18446744073709551615} {}`, `{"value":18446744073709551615} warning`} {
		runner.output = output
		var value map[string]any
		if err := s.read(context.Background(), id, "fixture", []string{"status", "--format", "json"}, &value); err == nil {
			t.Fatalf("accepted trailing data: %s", output)
		}
	}
	runner.output = "{\"value\":18446744073709551615}\n\t"
	var value map[string]any
	if err := s.read(context.Background(), id, "fixture", []string{"status", "--format", "json"}, &value); err != nil {
		t.Fatal(err)
	}
	if value["value"] != json.Number("18446744073709551615") {
		t.Fatalf("precision lost: %+v", value)
	}
	runner.output = `[{"seq":0,"stamp":"2026-10-03T00:00:00Z","message":""}]`
	if _, err := s.Logs(context.Background(), id, "cluster", "debug", 10); err != nil {
		t.Fatalf("zero sequence rejected: %v", err)
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

func TestLocalizedManagerConfigurationMetadata(t *testing.T) {
	service, runner, id := testInspection(t)
	const name = "mgr/dashboard/node-a.x1/server_port"
	runner.output = `{"name":"mgr/dashboard/server_port","type":"int","default":8080}`
	option, err := service.ConfigurationOption(context.Background(), id, name)
	if err != nil || option["name"] != name || option["metadata_name"] != "mgr/dashboard/server_port" || option["default"] != "8080" {
		t.Fatalf("localized metadata=%v err=%v", option, err)
	}
	if !reflect.DeepEqual(runner.specs[0].Args, []string{"config", "help", name, "--format", "json"}) {
		t.Fatal("localized help target lost")
	}
	runner.output = `{"name":"mgr/dashboard/other"}`
	if _, err := service.ConfigurationOption(context.Background(), id, name); err == nil {
		t.Fatal("unrelated metadata accepted")
	}
	for _, invalid := range []string{"mgr/dashboard//server_port", "mgr/dashboard/../server_port", "mgr/dashboard/-flag/server_port", "mgr/dashboard/node/extra/server_port", "mgr/dashboard/node a/server_port"} {
		before := len(runner.specs)
		if _, err := service.ConfigurationOption(context.Background(), id, invalid); err == nil || len(runner.specs) != before {
			t.Fatalf("invalid localized key executed: %s", invalid)
		}
	}
}

func TestConfigurationMetadataPreservesNumericText(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.output = `{"name":"test","default":18446744073709551615,"daemon_default":9007199254740993,"min":-9223372036854775808,"max":1.234567890123456789e20}`
	option, err := service.ConfigurationOption(context.Background(), id, "test")
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"default": "18446744073709551615", "daemon_default": "9007199254740993", "min": "-9223372036854775808", "max": "1.234567890123456789e20"} {
		if option[key] != want {
			t.Fatalf("%s = %#v", key, option[key])
		}
	}
	runner.output = `{"name":"test","default":false,"daemon_default":"","min":0,"max":"4G"}`
	option, err = service.ConfigurationOption(context.Background(), id, "test")
	if err != nil || option["default"] != false || option["daemon_default"] != "" || option["min"] != "0" || option["max"] != "4G" {
		t.Fatalf("option=%v err=%v", option, err)
	}
}

func TestConfigurationMetadataRejectsMalformedFields(t *testing.T) {
	service, runner, id := testInspection(t)
	for _, fields := range []string{
		`"can_update_at_runtime":"false"`, `"can_update_at_runtime":null`,
		`"type":1`, `"level":{}`, `"desc":false`, `"long_desc":[]`,
		`"enum_values":[1]`, `"tags":null`, `"services":"osd"`,
		`"see_also":[{}]`, `"flags":[true]`,
		`"default":{}`, `"daemon_default":[]`, `"min":null`, `"max":[]`,
	} {
		runner.output = `{"name":"test",` + fields + `}`
		if _, err := service.ConfigurationOption(context.Background(), id, "test"); err == nil {
			t.Fatalf("accepted %s", fields)
		}
	}
	for _, fields := range []string{
		`"can_update_at_runtime":false,"default":false,"min":"","max":0`,
		`"enum_values":[],"tags":["rados"],"services":["osd"],"flags":["runtime"],"see_also":[]`,
		`"default":18446744073709551615,"daemon_default":"4G"`,
	} {
		runner.output = `{"name":"test",` + fields + `}`
		if _, err := service.ConfigurationOption(context.Background(), id, "test"); err != nil {
			t.Fatal(fields, err)
		}
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
		if tc.section == "metadata" {
			runner.output = `{"id":0,"hostname":"node1","osd_objectstore":"bluestore"}`
		}
		result, err := service.OSDInspection(context.Background(), id, "0", tc.section)
		if err != nil || (tc.section == "histogram" && result["osd"] == nil) || (tc.section == "metadata" && result["hostname"] != "node1") {
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

func TestOSDMetadataRequiresMatchingNativeIdentity(t *testing.T) {
	service, runner, id := testInspection(t)
	for _, output := range []string{`{}`, `{"hostname":"node1"}`, `{"id":1}`, `{"id":"0"}`, `{"id":null}`, `{"id":false}`, `{"id":0.0}`, `{"id":0e0}`, `{"id":-1}`, `[{"id":0}]`} {
		runner.output = output
		result, err := service.OSDInspection(context.Background(), id, "0", "metadata")
		if err == nil || result != nil {
			t.Fatalf("accepted mismatched metadata %s: %v", output, result)
		}
	}
	runner.output = `{"id":42,"hostname":"node42","mem_total_kb":"18446744073709551615","bluefs":"1","custom":"kept"}`
	result, err := service.OSDInspection(context.Background(), id, "42", "metadata")
	if err != nil || result["hostname"] != "node42" || result["mem_total_kb"] != "18446744073709551615" || result["custom"] != "kept" {
		t.Fatalf("matching metadata not preserved: %v %v", result, err)
	}
	if !reflect.DeepEqual(runner.specs[len(runner.specs)-1].Args, []string{"osd", "metadata", "42", "--format", "json"}) {
		t.Fatal("wrong native target")
	}
}

func TestOSDDeviceInspection(t *testing.T) {
	service, runner, id := testInspection(t)
	for _, raw := range []string{`[]`, `[{"devid":"disk-1","location":[{"host":"node-a","dev":"sda"}],"daemons":["osd.0"]}]`} {
		runner.output = raw
		result, err := service.OSDInspection(context.Background(), id, "0", "devices")
		if err != nil || result["devices"] == nil {
			t.Fatalf("result=%v err=%v", result, err)
		}
		spec := runner.specs[len(runner.specs)-1]
		if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"device", "ls-by-daemon", "osd.0", "--format", "json"}) {
			t.Fatalf("spec=%+v", spec)
		}
	}
	for _, raw := range []string{`null`, `{}`, `[null]`, `[{}]`, `[{"devid":0}]`, `[{"devid":" "}]`} {
		runner.output = raw
		if _, err := service.OSDInspection(context.Background(), id, "0", "devices"); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	runner.fail = true
	if _, err := service.OSDInspection(context.Background(), id, "0", "devices"); err == nil {
		t.Fatal("command failure accepted")
	}
}

func TestOSDSMARTInspection(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.output = `{"disk-1":{"smart_status":{"passed":false},"counter":18446744073709551615,"password":"secret-fixture","nested":[9007199254740993]}}`
	result, err := service.OSDInspection(context.Background(), id, "12", "smart")
	if err != nil {
		t.Fatal(err)
	}
	report := result["disk-1"].(map[string]any)
	if report["counter"] != "18446744073709551615" || report["nested"].([]any)[0] != "9007199254740993" || report["smart_status"].(map[string]any)["passed"] != false || report["password"] == "secret-fixture" {
		t.Fatalf("report=%v", report)
	}
	spec := runner.specs[len(runner.specs)-1]
	if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"device", "query-daemon-health-metrics", "osd.12", "--format", "json"}) {
		t.Fatalf("spec=%+v", spec)
	}
	runner.output = `{}`
	if result, err := service.OSDInspection(context.Background(), id, "12", "smart"); err != nil || len(result) != 0 {
		t.Fatalf("result=%v err=%v", result, err)
	}
	for _, raw := range []string{`null`, `[]`, `{} {}`} {
		runner.output = raw
		if _, err := service.OSDInspection(context.Background(), id, "12", "smart"); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	runner.fail = true
	if _, err := service.OSDInspection(context.Background(), id, "12", "smart"); err == nil {
		t.Fatal("failed command accepted")
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
