package operation

import (
	"context"
	"sync"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const operationTestKey = "0123456789abcdefghijklmnopqrstuv"

type dispatcherFake struct {
	mu      sync.Mutex
	entered chan ExecutionRequest
	release chan struct{}
	active  map[uint64]int
	max     map[uint64]int
}

func (f *dispatcherFake) Execute(ctx context.Context, request ExecutionRequest) (cephdomain.ActionResult, error) {
	f.mu.Lock()
	f.active[request.ClusterID]++
	if f.active[request.ClusterID] > f.max[request.ClusterID] {
		f.max[request.ClusterID] = f.active[request.ClusterID]
	}
	f.mu.Unlock()
	f.entered <- request
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
			return cephdomain.ActionResult{}, ctx.Err()
		}
	}
	f.mu.Lock()
	f.active[request.ClusterID]--
	f.mu.Unlock()
	return cephdomain.ActionResult{Details: map[string]any{"executed": true}}, nil
}

func TestEnqueueEncryptsParametersAndIsIdempotent(t *testing.T) {
	db, clusterID := operationServiceDatabase(t)
	service := New(func() *store.Database { return db }, operationTestKey, &dispatcherFake{}, Options{})
	request := EnqueueRequest{
		ClusterID: clusterID, RequestID: "request", IdempotencyKey: "idempotent",
		Action: "pool.create", ResourceKind: "pool", ResourceKey: "pool-a", Risk: "medium",
		LockKey: "pool/pool-a", Parameters: map[string]any{"name": "pool-a", "secret_key": "secret"},
	}
	first, err := service.Enqueue(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Enqueue(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("idempotent operation ids = %d and %d", first.ID, second.ID)
	}
	if first.ParametersCiphertext == "" || first.ParametersCiphertext == `{"name":"pool-a","secret_key":"secret"}` {
		t.Fatalf("operation parameters were not encrypted: %q", first.ParametersCiphertext)
	}
	plaintext, err := security.Decrypt(first.ParametersCiphertext, operationTestKey)
	if err != nil || string(plaintext) != `{"name":"pool-a","secret_key":"secret"}` {
		t.Fatalf("decrypted parameters = %q, err=%v", plaintext, err)
	}
}

func TestWorkersSerializeOperationsForTheSameCluster(t *testing.T) {
	db, clusterID := operationServiceDatabase(t)
	dispatcher := &dispatcherFake{
		entered: make(chan ExecutionRequest, 2), release: make(chan struct{}),
		active: map[uint64]int{}, max: map[uint64]int{},
	}
	service := New(func() *store.Database { return db }, operationTestKey, dispatcher, Options{Workers: 2, PollInterval: 10 * time.Millisecond})
	for _, name := range []string{"pool-a", "pool-b"} {
		if _, err := service.Enqueue(context.Background(), EnqueueRequest{
			ClusterID: clusterID, RequestID: name, IdempotencyKey: name, Action: "pool.create",
			ResourceKind: "pool", ResourceKey: name, Risk: "medium", LockKey: "pool/" + name,
			Parameters: map[string]any{"name": name},
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(service.Stop)
	select {
	case <-dispatcher.entered:
	case <-time.After(time.Second):
		t.Fatal("first operation was not dispatched")
	}
	select {
	case request := <-dispatcher.entered:
		t.Fatalf("same-cluster operation ran concurrently: %#v", request)
	case <-time.After(75 * time.Millisecond):
	}
	close(dispatcher.release)
	select {
	case <-dispatcher.entered:
	case <-time.After(time.Second):
		t.Fatal("second operation was not dispatched after the first completed")
	}
	waitForOperationStatus(t, db, 1, store.OperationSucceeded)
	waitForOperationStatus(t, db, 2, store.OperationSucceeded)
	dispatcher.mu.Lock()
	maxActive := dispatcher.max[clusterID]
	dispatcher.mu.Unlock()
	if maxActive != 1 {
		t.Fatalf("maximum same-cluster concurrency = %d, want 1", maxActive)
	}
}

func waitForOperationStatus(t *testing.T, db *store.Database, id uint64, status string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		row, err := db.FindOperation(context.Background(), id)
		if err == nil && row.Status == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	row, err := db.FindOperation(context.Background(), id)
	t.Fatalf("operation %d status = %q, err=%v, want %q", id, row.Status, err, status)
}

func operationServiceDatabase(t *testing.T) (*store.Database, uint64) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: operationTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "operation-service.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "operation-service", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	return db, cluster.ID
}
