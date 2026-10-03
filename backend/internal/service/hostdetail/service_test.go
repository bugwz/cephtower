package hostdetail

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/store"
)

const testEncryptionKey = "0123456789abcdefghijklmnopqrstuv"

type fakeExecutor struct {
	args    [][]string
	outputs [][]byte
	failAt  map[int]error
}

func (f *fakeExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	f.args = append(f.args, append([]string(nil), spec.Args...))
	if err := f.failAt[len(f.args)]; err != nil {
		return executor.CommandResult{}, err
	}
	output := f.outputs[0]
	f.outputs = f.outputs[1:]
	return executor.CommandResult{Stdout: output}, nil
}

func TestSMARTDoesNotHideFailedDaemonQueries(t *testing.T) {
	for _, partial := range []bool{false, true} {
		service, runner, id := testService(t,
			[]byte(`[{"devid":"disk-1","daemons":["osd.2","osd.1"]}]`),
			[]byte(`{"disk-1":{"smart_status":{"passed":true}}}`),
		)
		failureCall := 2
		if partial {
			failureCall = 3
		}
		runner.failAt = map[int]error{failureCall: errors.New("daemon query failed")}
		result, err := service.SMART(context.Background(), id, "node-1")
		if err == nil || result != nil || len(runner.args) != failureCall {
			t.Fatalf("failed query hidden: %v %+v", err, result)
		}
		if runner.args[1][2] != "osd.1" {
			t.Fatal("daemon query order is not deterministic")
		}
	}
}

func TestSMARTRejectsNullNativeResponses(t *testing.T) {
	service, _, id := testService(t, []byte(`null`))
	if result, err := service.SMART(context.Background(), id, "node-1"); err == nil || result != nil {
		t.Fatal("null device inventory accepted")
	}
	for _, output := range []string{`null`, `[]`, `{bad`, `{} {}`} {
		service, _, id = testService(t, []byte(`[{"devid":"disk-1","daemons":["osd.1"]}]`), []byte(output))
		if result, err := service.SMART(context.Background(), id, "node-1"); err == nil || result != nil {
			t.Fatalf("accepted %s", output)
		}
	}
	service, _, id = testService(t, []byte(`[]`))
	if result, err := service.SMART(context.Background(), id, "node-1"); err != nil || result == nil || len(result) != 0 {
		t.Fatal("valid empty inventory rejected")
	}
}

func TestDevicesUsesCephDeviceListByHost(t *testing.T) {
	service, runner, clusterID := testService(t, []byte(`[{"devid":"disk-1","life_expectancy_stamp":"2026-10-03T01:02:03Z"}]`))
	devices, err := service.Devices(context.Background(), clusterID, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0]["devid"] != "disk-1" {
		t.Fatalf("devices = %#v", devices)
	}
	if devices[0]["life_expectancy_stamp"] != "2026-10-03T01:02:03Z" {
		t.Fatal("device prediction timestamp was lost")
	}
	want := []string{"device", "ls-by-host", "node-1", "--format", "json"}
	if !reflect.DeepEqual(runner.args[0], want) {
		t.Fatalf("args = %#v, want %#v", runner.args[0], want)
	}
}

func TestSMARTQueriesAssociatedDaemon(t *testing.T) {
	service, runner, clusterID := testService(t,
		[]byte(`[{"devid":"disk-1","daemons":["osd.1"]}]`),
		[]byte(`{"disk-1":{"smart_status":{"passed":true}}}`),
	)
	payload, err := service.SMART(context.Background(), clusterID, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["disk-1"]; !ok {
		t.Fatalf("SMART payload = %#v", payload)
	}
	want := []string{"device", "query-daemon-health-metrics", "osd.1", "--format", "json"}
	if !reflect.DeepEqual(runner.args[1], want) {
		t.Fatalf("args = %#v, want %#v", runner.args[1], want)
	}
}

func TestSMARTPreservesDetailedCountersAndRedactsSecrets(t *testing.T) {
	service, _, id := testService(t,
		[]byte(`[{"devid":"disk-1","daemons":["osd.1"]}]`),
		[]byte(`{"disk-1":{"smart_status":{"passed":true},"ata_smart_attributes":{"table":[{"id":9,"raw":{"value":9007199254740993}}]},"scsi_error_counter_log":{"read":{"total_uncorrected_errors":18446744073709551615}},"nvme_smart_health_information_log":{"percentage_used":0,"data_units_written":18446744073709551615},"vendor":{"password":"fixture-secret"}}}`),
	)
	result, err := service.SMART(context.Background(), id, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"value":"9007199254740993"`, `"total_uncorrected_errors":"18446744073709551615"`, `"percentage_used":"0"`, `"passed":true`, `[REDACTED]`} {
		if !strings.Contains(string(raw), expected) {
			t.Fatalf("missing %s in %s", expected, raw)
		}
	}
	if strings.Contains(string(raw), "fixture-secret") {
		t.Fatal("SMART detail leaked credentials")
	}
}

func TestDevicesRejectsAmbiguousIdentities(t *testing.T) {
	for _, output := range []string{`[null]`, `[{}]`, `[{"devid":1}]`, `[{"devid":" "}]`, `[{"devid":"disk-1"},{"devid":"disk-1"}]`} {
		t.Run(output, func(t *testing.T) {
			service, runner, id := testService(t, []byte(output))
			if result, err := service.SMART(context.Background(), id, "node-1"); err == nil || result != nil {
				t.Fatal("accepted ambiguous inventory")
			}
			if len(runner.args) != 1 {
				t.Fatal("queried SMART with invalid inventory")
			}
		})
	}
}

func TestSMARTScopesReportsAndRetainsMissingDevices(t *testing.T) {
	service, _, id := testService(t,
		[]byte(`[{"devid":"disk-1","daemons":["osd.1"]},{"devid":"disk-2"}]`),
		[]byte(`{"disk-1":{"smart_status":{"passed":true}},"foreign-disk":{"smart_status":{"passed":true}}}`),
	)
	result, err := service.SMART(context.Background(), id, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 || result["foreign-disk"] != nil || result["disk-1"] == nil {
		t.Fatalf("unscoped reports: %#v", result)
	}
	missing, ok := result["disk-2"].(map[string]any)
	if !ok || missing["error"] == nil || missing["smart_status"] != nil {
		t.Fatalf("missing device incorrectly reported: %#v", result)
	}
	for _, inventory := range []string{`[{"devid":"disk-1"}]`, `[{"devid":"disk-1","daemons":["osd.1"]}]`} {
		service, _, id := testService(t, []byte(inventory), []byte(`{}`))
		result, err := service.SMART(context.Background(), id, "node-1")
		if err != nil || len(result) != 1 {
			t.Fatalf("missing device disappeared: %#v, %v", result, err)
		}
		if row, ok := result["disk-1"].(map[string]any); !ok || row["error"] == nil {
			t.Fatalf("missing report not marked unknown: %#v", result)
		}
	}
}

func testService(t *testing.T, outputs ...[]byte) (*Service, *fakeExecutor, uint64) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: testEncryptionKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "hostdetail.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	key, err := security.Encrypt([]byte("secret"), testEncryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: key, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	clusters := clusterservice.New(func() *store.Database { return db }, testEncryptionKey, nilClusterProvider{})
	runner := &fakeExecutor{outputs: outputs}
	return New(clusters, runner), runner, cluster.ID
}

type nilClusterProvider struct{}

func (nilClusterProvider) Probe(context.Context, cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	return cephprovider.ProbeResult{}, nil
}
