package mutation

import (
	"context"
	"errors"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

const planTestKey = "0123456789abcdefghijklmnopqrstuv"

type planExecutor struct {
	t      *testing.T
	failID string
}

func (f planExecutor) Run(_ context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if access.ClientKey != "plain-ceph-key" {
		f.t.Fatal("plan checker did not use decrypted cluster access")
	}
	if spec.Mutating {
		f.t.Fatalf("plan command %s was marked mutating", spec.ID)
	}
	if spec.ID == f.failID {
		return executor.CommandResult{}, errors.New("precondition denied")
	}
	stdout := []byte(`{}`)
	if spec.ID == "plan.pool_rbd_images" || spec.ID == "plan.device_inventory" {
		stdout = []byte(`[]`)
	}
	return executor.CommandResult{Stdout: stdout}, nil
}

func planService(t *testing.T, failID string) (*Service, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: planTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/plan.db"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	encrypted, err := security.Encrypt([]byte("plain-ceph-key"), planTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	clusters := clusterservice.New(func() *store.Database { return db }, planTestKey, nil, nil)
	return New(clusters, planExecutor{t: t, failID: failID}, planTestKey), cluster
}

func TestCheckPlanBlocksUnsafeOSDRemoval(t *testing.T) {
	service, cluster := planService(t, "plan.osd_safe_to_destroy")
	blockers, warnings, err := service.CheckPlan(context.Background(), operationservice.PlanRequest{ClusterID: cluster.ID, Action: "osd.delete", ResourceKind: "osd", ResourceKey: "osd/7"})
	if err != nil || len(blockers) != 1 || len(warnings) != 0 {
		t.Fatalf("blockers=%v warnings=%v err=%v", blockers, warnings, err)
	}
}

func TestCheckPlanAllowsSafeOSDRemoval(t *testing.T) {
	service, cluster := planService(t, "")
	blockers, warnings, err := service.CheckPlan(context.Background(), operationservice.PlanRequest{ClusterID: cluster.ID, Action: "osd.delete", ResourceKind: "osd", ResourceKey: "osd/7"})
	if err != nil || len(blockers) != 0 || len(warnings) != 0 {
		t.Fatalf("blockers=%v warnings=%v err=%v", blockers, warnings, err)
	}
}
