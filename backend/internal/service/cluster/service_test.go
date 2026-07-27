package cluster

import (
	"context"
	"errors"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/security"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

const clusterTestKey = "0123456789abcdefghijklmnopqrstuv"

type probeProvider struct{ err error }

func (p probeProvider) Probe(_ context.Context, access cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	if p.err != nil {
		return cephprovider.ProbeResult{}, p.err
	}
	return cephprovider.ProbeResult{FSID: "00000000-0000-0000-0000-000000000001", Version: "ceph version 20.2.2", Capabilities: []cephprovider.Capability{{Name: "ceph_cli", Supported: true}}, Status: map[string]any{"client": access.ClientUsername}}, nil
}

func clusterTestServices(t *testing.T, provider cephprovider.ClusterProvider) (*Service, *store.Database, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: clusterTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/cluster.db"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	encrypted, err := security.Encrypt([]byte("old-secret"), clusterTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	row := store.CephCluster{Name: "before", MonitorAddresses: "mon.example.test:6789", ClientUsername: "client.before", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &row); err != nil {
		t.Fatal(err)
	}
	operations := operationservice.New(func() *store.Database { return db }, 1, clusterTestKey)
	return New(func() *store.Database { return db }, clusterTestKey, operations, provider), db, row
}

func TestUpdatePersistsOnlyAfterSuccessfulProbe(t *testing.T) {
	service, db, row := clusterTestServices(t, probeProvider{})
	name, username, clientKey := "after", "client.after", "new-secret"
	operation, err := service.Update(context.Background(), row.ID, UpdateInput{Name: &name, ClientUsername: &username, ClientKey: &clientKey}, nil, "tester", "request", "")
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := db.FindCluster(context.Background(), row.ID)
	if stored.Name != "before" || stored.ClientUsername != "client.before" {
		t.Fatalf("cluster changed before worker execution: %#v", stored)
	}
	if _, err := service.probeOperation(context.Background(), operation); err != nil {
		t.Fatal(err)
	}
	stored, _ = db.FindCluster(context.Background(), row.ID)
	if stored.Name != name || stored.ClientUsername != username {
		t.Fatalf("cluster was not updated: %#v", stored)
	}
	plain, err := security.Decrypt(stored.ClientKey, clusterTestKey)
	if err != nil || string(plain) != clientKey {
		t.Fatalf("client key = %q, err=%v", plain, err)
	}
}

func TestUpdateProbeFailureKeepsPreviousAccess(t *testing.T) {
	service, db, row := clusterTestServices(t, probeProvider{err: errors.New("unreachable")})
	name := "after"
	operation, err := service.Update(context.Background(), row.ID, UpdateInput{Name: &name}, nil, "tester", "request", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.probeOperation(context.Background(), operation); err == nil {
		t.Fatal("failed probe returned success")
	}
	stored, _ := db.FindCluster(context.Background(), row.ID)
	if stored.Name != "before" {
		t.Fatalf("failed update changed cluster: %#v", stored)
	}
}

func TestClusterServiceLeavesRefreshForTheReconciler(t *testing.T) {
	_, db, _ := clusterTestServices(t, probeProvider{})
	operations := operationservice.New(func() *store.Database { return db }, 1, clusterTestKey)
	_ = New(func() *store.Database { return db }, clusterTestKey, operations, probeProvider{})
	if err := operations.Register("cluster.refresh", func(context.Context, store.CephOperation) (cephdomain.OperationResult, error) {
		return cephdomain.OperationResult{}, nil
	}); err != nil {
		t.Fatalf("register reconciler refresh handler: %v", err)
	}
}
