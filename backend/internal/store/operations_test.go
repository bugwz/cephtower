package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"gorm.io/gorm"
)

func TestOperationLifecycle(t *testing.T) {
	db, clusterID := operationTestDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	idempotencyKey := "request-1"
	operation := CephOperation{
		ClusterID: clusterID, RequestID: "request-id", IdempotencyKey: &idempotencyKey,
		Action: "pool.create", ResourceKind: "pool", ResourceKey: "pool-a", Risk: "medium",
		LockKey: "pool/pool-a", Status: OperationQueued, ParametersCiphertext: "ciphertext",
		MaxAttempts: 2, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.CreateOperation(ctx, &operation); err != nil {
		t.Fatal(err)
	}
	stored, err := db.FindOperationByIdempotencyKey(ctx, clusterID, idempotencyKey)
	if err != nil || stored.ID != operation.ID {
		t.Fatalf("idempotent operation = %#v, err=%v", stored, err)
	}

	claimed, err := db.ClaimNextOperation(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if claimed.Status != OperationRunning || claimed.Attempts != 1 || claimed.StartedAt == nil {
		t.Fatalf("claimed operation = %#v", claimed)
	}
	if _, err := db.ClaimNextOperation(ctx, now); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("second claim error = %v, want record not found", err)
	}

	nextAttempt := now.Add(time.Minute)
	if err := db.RequeueOperation(ctx, operation.ID, "temporary", "retry later", nextAttempt, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimNextOperation(ctx, now); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("early retry claim error = %v, want record not found", err)
	}
	claimed, err = db.ClaimNextOperation(ctx, nextAttempt)
	if err != nil || claimed.Attempts != 2 {
		t.Fatalf("retry claim = %#v, err=%v", claimed, err)
	}
	if err := db.CompleteOperation(ctx, operation.ID, `{"ok":true}`, nextAttempt.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	completed, err := db.FindOperation(ctx, operation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != OperationSucceeded || completed.ResultJSON == nil || completed.FinishedAt == nil {
		t.Fatalf("completed operation = %#v", completed)
	}
}

func TestRecoverRunningOperations(t *testing.T) {
	db, clusterID := operationTestDatabase(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for _, status := range []string{OperationRunning, OperationSucceeded} {
		row := CephOperation{
			ClusterID: clusterID, RequestID: status, Action: "pool.create", ResourceKind: "pool",
			ResourceKey: status, Risk: "medium", LockKey: "pool/" + status, Status: status,
			ParametersCiphertext: "ciphertext", MaxAttempts: 1, CreatedAt: now, UpdatedAt: now,
		}
		if err := db.CreateOperation(ctx, &row); err != nil {
			t.Fatal(err)
		}
	}
	count, err := db.RecoverRunningOperations(ctx, now.Add(time.Minute))
	if err != nil || count != 1 {
		t.Fatalf("recovered count=%d err=%v", count, err)
	}
	failed, err := db.ListOperations(ctx, OperationFilter{Status: OperationFailed})
	if err != nil || len(failed) != 1 || failed[0].ErrorCode == nil || *failed[0].ErrorCode != "operation_interrupted" {
		t.Fatalf("failed operations=%#v err=%v", failed, err)
	}
}

func TestPruneCompletedOperationsPreservesActiveAndRecentRows(t *testing.T) {
	db, clusterID := operationTestDatabase(t)
	now := time.Now().UTC()
	oldFinished := now.Add(-100 * 24 * time.Hour)
	recentFinished := now.Add(-time.Hour)
	rows := []CephOperation{
		{ClusterID: clusterID, RequestID: "old", Action: "pool.create", ResourceKind: "pool", Status: OperationSucceeded, ParametersCiphertext: "cipher", MaxAttempts: 1, FinishedAt: &oldFinished, CreatedAt: oldFinished, UpdatedAt: oldFinished},
		{ClusterID: clusterID, RequestID: "active", Action: "pool.create", ResourceKind: "pool", Status: OperationQueued, ParametersCiphertext: "cipher", MaxAttempts: 1, CreatedAt: oldFinished, UpdatedAt: oldFinished},
		{ClusterID: clusterID, RequestID: "recent", Action: "pool.create", ResourceKind: "pool", Status: OperationFailed, ParametersCiphertext: "cipher", MaxAttempts: 1, FinishedAt: &recentFinished, CreatedAt: recentFinished, UpdatedAt: recentFinished},
	}
	for index := range rows {
		if err := db.CreateOperation(context.Background(), &rows[index]); err != nil {
			t.Fatal(err)
		}
	}
	deleted, err := db.PruneCompletedOperations(context.Background(), now.Add(-90*24*time.Hour))
	if err != nil || deleted != 1 {
		t.Fatalf("deleted=%d err=%v", deleted, err)
	}
	for _, row := range rows[1:] {
		if _, err := db.FindOperation(context.Background(), row.ID); err != nil {
			t.Fatalf("operation %d was pruned: %v", row.ID, err)
		}
	}
}

func operationTestDatabase(t *testing.T) (*Database, uint64) {
	t.Helper()
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "operations.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(db) })
	now := time.Now().UTC()
	cluster := CephCluster{Name: "operations", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	return db, cluster.ID
}
