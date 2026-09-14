package mutation

import (
	"context"
	"fmt"
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

func TestCephUserCommands(t *testing.T) {
	for _, tt := range []struct {
		action string
		params map[string]any
		args   []string
	}{
		{"ceph_user.create", map[string]any{"entity": "client.backup", "caps": map[string]any{"mon": "allow r", "osd": "profile rbd pool=data"}}, []string{"auth", "add", "client.backup", "mon", "allow r", "osd", "profile rbd pool=data"}},
		{"ceph_user.update", map[string]any{"entity": "client.backup", "caps": map[string]any{"mds": `allow rw path="/shared folder"`}}, []string{"auth", "caps", "client.backup", "mds", `allow rw path="/shared folder"`}},
		{"ceph_user.delete", map[string]any{"entity": "client.backup"}, []string{"auth", "rm", "client.backup"}},
	} {
		t.Run(tt.action, func(t *testing.T) {
			cmd, err := build(Request{Action: tt.action}, tt.params)
			if err != nil || !reflect.DeepEqual(cmd.args, tt.args) {
				t.Fatalf("args=%v, err=%v", cmd.args, err)
			}
			if !Supports(tt.action) {
				t.Fatal("action not supported")
			}
		})
	}
	for _, params := range []map[string]any{
		{"entity": "--help", "caps": map[string]any{"mon": "allow r"}},
		{"entity": "client.a;id", "caps": map[string]any{"mon": "allow r"}},
		{"entity": "client.a", "caps": map[string]any{}},
		{"entity": "client.a", "caps": map[string]any{"unknown": "allow *"}},
		{"entity": "client.a", "caps": map[string]any{"mon": true}},
		{"entity": "client.a", "caps": map[string]any{"mon": "--out-file=/tmp/unwanted"}},
		{"entity": "client.a", "caps": map[string]any{"mon": "allow r\x00"}},
	} {
		if _, err := build(Request{Action: "ceph_user.update"}, params); err == nil {
			t.Fatalf("accepted %v", params)
		}
	}
}

func TestCephUserImportUsesStdin(t *testing.T) {
	keyring := "[client.backup]\n key = secret-example\n caps mon = \"allow r\"\n"
	cmd, err := build(Request{Action: "ceph_user.import"}, map[string]any{"keyring": keyring})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cmd.args, []string{"auth", "import", "-i", "-"}) || string(cmd.stdin) != keyring {
		t.Fatalf("import command incorrect")
	}
	if strings.Contains(strings.Join(cmd.args, " "), "secret-example") {
		t.Fatal("secret in arguments")
	}
	redacted, err := security.RedactJSON(map[string]any{"keyring": keyring})
	if err != nil || strings.Contains(fmt.Sprint(redacted), "secret-example") {
		t.Fatal("keyring not redacted")
	}
	for _, keyring := range []string{"", "\x00", strings.Repeat("x", (256<<10)+1)} {
		if _, err := build(Request{Action: "ceph_user.import"}, map[string]any{"keyring": keyring}); err == nil {
			t.Fatal("invalid keyring accepted")
		}
	}
}

type cephUserExecutor struct {
	specs []executor.CommandSpec
	fail  bool
}

func (e *cephUserExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if e.fail {
		return executor.CommandResult{}, fmt.Errorf("key = should-never-appear")
	}
	return executor.CommandResult{Stdout: []byte("[" + spec.Args[2] + "]\n key = synthetic-secret\n")}, nil
}

type cephUserProbe struct{}

func (cephUserProbe) Probe(context.Context, cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	return cephprovider.ProbeResult{}, nil
}

func newCephUserService(t *testing.T) (*Service, *cephUserExecutor, uint64) {
	t.Helper()
	const encryptionKey = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: encryptionKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "ceph-user.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	secret, err := security.Encrypt([]byte("fixture-key"), encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: secret, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	service := clusterservice.New(func() *store.Database { return db }, encryptionKey, cephUserProbe{})
	runner := &cephUserExecutor{}
	return New(service, runner), runner, cluster.ID
}

func TestCephUserExportIsExplicitReadAndHidesErrors(t *testing.T) {
	service, runner, id := newCephUserService(t)
	for _, entities := range [][]string{nil, {""}, {"--help"}, {"client.a", "client.a"}} {
		if _, err := service.ExportCephUsers(context.Background(), id, entities); err == nil {
			t.Fatalf("accepted invalid selection %v", entities)
		}
	}
	if len(runner.specs) != 0 {
		t.Fatal("invalid export executed")
	}
	value, err := service.ExportCephUsers(context.Background(), id, []string{"client.a", "client.b"})
	if err != nil || !strings.Contains(value, "[client.a]") || !strings.Contains(value, "[client.b]") {
		t.Fatalf("export failed: %v", err)
	}
	for i, spec := range runner.specs {
		if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"auth", "export", []string{"client.a", "client.b"}[i]}) {
			t.Fatalf("wrong export spec: %+v", spec)
		}
	}
	runner.fail = true
	if _, err := service.ExportCephUsers(context.Background(), id, []string{"client.a"}); err == nil || strings.Contains(err.Error(), "should-never-appear") {
		t.Fatalf("unsafe export failure: %v", err)
	}
	if _, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "ceph_user.import", Parameters: map[string]any{"keyring": "[client.a]\nkey = foo"}}); err == nil || strings.Contains(err.Error(), "should-never-appear") {
		t.Fatalf("unsafe import failure: %v", err)
	}
}
