package hostdetail

import (
	"context"
	"reflect"
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
}

func (f *fakeExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	f.args = append(f.args, append([]string(nil), spec.Args...))
	output := f.outputs[0]
	f.outputs = f.outputs[1:]
	return executor.CommandResult{Stdout: output}, nil
}

func TestDevicesUsesCephDeviceListByHost(t *testing.T) {
	service, runner, clusterID := testService(t, []byte(`[{"devid":"disk-1"}]`))
	devices, err := service.Devices(context.Background(), clusterID, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 || devices[0]["devid"] != "disk-1" {
		t.Fatalf("devices = %#v", devices)
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
