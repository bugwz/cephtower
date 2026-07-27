package operation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const operationTestKey = "0123456789abcdefghijklmnopqrstuv"

func operationDatabase(t *testing.T) (*store.Database, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: operationTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/operation.db"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "test", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	return db, cluster
}
func TestEnqueueIsIdempotent(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	request := Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, RequestID: "request", Action: "pool.create", ResourceKind: "pool", ResourceKey: "pool/test", Risk: cephdomain.RiskMedium, IdempotencyKey: "same", Parameters: map[string]any{"name": "test"}}
	first, err := service.Enqueue(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Enqueue(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("operations differ: %s %s", first.ID, second.ID)
	}
}

func TestIdempotencyScopeUsesActorAndClusterValues(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	actor1, actor2 := uint64(42), uint64(42)
	now := time.Now().UTC()
	if err := db.CreateUser(context.Background(), &store.User{ID: actor1, Username: "actor", Password: "encrypted", Status: "active", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	cluster1, cluster2 := cluster.ID, cluster.ID
	request := Request{ClusterID: &cluster1, ActorUserID: &actor1, ClusterName: cluster.Name, Action: "pool.create", ResourceKind: "pool", ResourceKey: "pool/test", Risk: cephdomain.RiskMedium, IdempotencyKey: "same", Parameters: map[string]any{"name": "test"}}
	first, err := service.Enqueue(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	request.ClusterID, request.ActorUserID = &cluster2, &actor2
	second, err := service.Enqueue(context.Background(), request)
	if err != nil || first.ID != second.ID {
		t.Fatalf("first=%s second=%s err=%v", first.ID, second.ID, err)
	}
}

func TestEnqueueEncryptsSensitiveRequestFields(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1, operationTestKey)
	row, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "iscsi_target.create", ResourceKind: "iscsi_target", ResourceKey: "iscsi/target/test", Risk: cephdomain.RiskMedium, Parameters: map[string]any{"username": "client", "password": "plain-password"}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(row.RequestJSON, "plain-password") || strings.Contains(row.RequestJSON, "[REDACTED]") {
		t.Fatalf("request JSON is not encrypted: %s", row.RequestJSON)
	}
	var stored any
	if err := json.Unmarshal([]byte(row.RequestJSON), &stored); err != nil {
		t.Fatal(err)
	}
	restored, err := security.UnprotectJSON(stored, operationTestKey)
	if err != nil {
		t.Fatal(err)
	}
	parameters := restored.(map[string]any)
	if parameters["password"] != "plain-password" {
		t.Fatalf("restored parameters = %#v", parameters)
	}
}

func TestEnqueueAuditStoresOnlyRedactedParametersAndAcceptedMetadata(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1, operationTestKey)
	generation := uint64(9)
	operation, err := service.Enqueue(context.Background(), Request{
		ClusterID:          &cluster.ID,
		ClusterName:        cluster.Name,
		RequestID:          "audit-request",
		Action:             "iscsi_target.create",
		ResourceKind:       "iscsi_target",
		ResourceKey:        "iscsi/target/test",
		ResourceGeneration: &generation,
		Risk:               cephdomain.RiskMedium,
		Parameters:         map[string]any{"username": "client", "password": "plain-password"},
	})
	if err != nil {
		t.Fatal(err)
	}
	audits, err := db.ListAuditEvents(context.Background(), cluster.ID, 10)
	if err != nil || len(audits) != 1 {
		t.Fatalf("audits=%d err=%v", len(audits), err)
	}
	audit := audits[0]
	if audit.OperationID == nil || *audit.OperationID != operation.ID || audit.HTTPStatus == nil || *audit.HTTPStatus != 202 {
		t.Fatalf("accepted audit metadata = %#v", audit)
	}
	if audit.BeforeGeneration == nil || *audit.BeforeGeneration != generation {
		t.Fatalf("before generation = %#v", audit.BeforeGeneration)
	}
	if audit.ParametersJSON == nil || strings.Contains(*audit.ParametersJSON, "plain-password") || !strings.Contains(*audit.ParametersJSON, "[REDACTED]") || !strings.Contains(*audit.ParametersJSON, "client") {
		t.Fatalf("audit parameters = %#v", audit.ParametersJSON)
	}
	filtered, err := db.ListAuditEventsFiltered(context.Background(), cluster.ID, store.AuditFilter{Action: "other.action", Limit: 10})
	if err != nil || len(filtered) != 0 {
		t.Fatalf("mismatched audit filter returned %d rows, err=%v", len(filtered), err)
	}
	filtered, err = db.ListAuditEventsFiltered(context.Background(), cluster.ID, store.AuditFilter{Action: "iscsi_target.create", ResourceKind: "iscsi_target", Limit: 10})
	if err != nil || len(filtered) != 1 {
		t.Fatalf("matching audit filter returned %d rows, err=%v", len(filtered), err)
	}
}
func TestHighRiskRequiresAndConsumesPlan(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	generation := uint64(7)
	request := Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: &generation, Risk: cephdomain.RiskHigh, Parameters: map[string]any{}}
	if _, err := service.Enqueue(context.Background(), request); err == nil {
		t.Fatal("high risk operation accepted without plan")
	}
	plan, err := service.CreatePlan(context.Background(), PlanRequest{ClusterID: cluster.ID, RequestID: "request", Action: request.Action, ResourceKind: request.ResourceKind, ResourceKey: request.ResourceKey, ResourceGeneration: generation, Risk: cephdomain.RiskHigh, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	request.PlanID = &plan.ID
	if _, err := service.Enqueue(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Enqueue(context.Background(), request); err == nil {
		t.Fatal("consumed plan was accepted twice")
	}
}

func TestServerSidePlanCheckerCanBlockConsumption(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	service.SetPlanChecker(func(context.Context, PlanRequest) ([]string, []string, error) {
		return []string{"resource is not safe to destroy"}, []string{"offline fixture warning"}, nil
	})
	plan, err := service.CreatePlan(context.Background(), PlanRequest{ClusterID: cluster.ID, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: 0, Risk: cephdomain.RiskHigh, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Status != "blocked" || !strings.Contains(plan.BlockersJSON, "not safe") || !strings.Contains(plan.WarningsJSON, "offline fixture") {
		t.Fatalf("plan = %#v", plan)
	}
	generation := uint64(0)
	_, err = service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: &generation, Risk: cephdomain.RiskHigh, PlanID: &plan.ID, Parameters: map[string]any{}})
	if err == nil {
		t.Fatal("blocked plan was consumed")
	}
}

func TestHighRiskPrecheckRunsAgainInWorker(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	checks := 0
	service.SetPlanChecker(func(context.Context, PlanRequest) ([]string, []string, error) {
		checks++
		if checks == 2 {
			return []string{"state changed after plan creation"}, nil, nil
		}
		return nil, nil, nil
	})
	service.SetFallback(func(context.Context, store.CephOperation) (cephdomain.OperationResult, error) {
		t.Fatal("mutation handler ran after the repeated pre-check was blocked")
		return cephdomain.OperationResult{}, nil
	})
	plan, err := service.CreatePlan(context.Background(), PlanRequest{ClusterID: cluster.ID, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: 0, Risk: cephdomain.RiskHigh, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	generation := uint64(0)
	operation, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: &generation, Risk: cephdomain.RiskHigh, PlanID: &plan.ID, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	service.runOne(context.Background())
	operation, err = service.Get(context.Background(), operation.ID)
	if err != nil {
		t.Fatal(err)
	}
	if checks != 2 || operation.Status != "failed" || operation.ErrorCode == nil || *operation.ErrorCode != "precondition_blocked" {
		t.Fatalf("checks=%d operation=%#v", checks, operation)
	}
}

func TestHighRiskPlanBindsParametersAndAllowsIdempotentReplay(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1, operationTestKey)
	generation := uint64(0)
	plannedParameters := map[string]any{"pool": "rbd", "expire": "7d"}
	plan, err := service.CreatePlan(context.Background(), PlanRequest{ClusterID: cluster.ID, Action: "rbd_trash.purge", ResourceKind: "rbd_trash", ResourceKey: "rbd/trash/purge", ResourceGeneration: generation, Risk: cephdomain.RiskHigh, Parameters: plannedParameters})
	if err != nil {
		t.Fatal(err)
	}
	request := Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "rbd_trash.purge", ResourceKind: "rbd_trash", ResourceKey: "rbd/trash/purge", ResourceGeneration: &generation, Risk: cephdomain.RiskHigh, PlanID: &plan.ID, IdempotencyKey: "purge-once", Parameters: map[string]any{"pool": "other", "expire": "7d"}}
	if _, err := service.Enqueue(context.Background(), request); err == nil || !strings.Contains(err.Error(), "parameters do not match") {
		t.Fatalf("mismatched parameters error = %v", err)
	}
	request.Parameters = plannedParameters
	first, err := service.Enqueue(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Enqueue(context.Background(), request)
	if err != nil || first.ID != second.ID {
		t.Fatalf("idempotent high-risk replay: first=%s second=%s err=%v", first.ID, second.ID, err)
	}
}
func TestPlanRejectsCurrentResourceGenerationConflict(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	now := time.Now().UTC()
	name := "test"
	if err := db.Insert(context.Background(), &store.CephResourceRecord{ClusterID: cluster.ID, Kind: "pool", NaturalKey: "test", Name: &name, Generation: 1, ResourceVersion: 7, Source: "fixture", ObservedAt: now, PayloadSchemaVersion: 1, PayloadJSON: `{}`, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	_, err := service.CreatePlan(context.Background(), PlanRequest{ClusterID: cluster.ID, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: 6, Risk: cephdomain.RiskHigh, Parameters: map[string]any{}})
	if err == nil {
		t.Fatal("stale generation was accepted")
	}
}

func TestResourceLookupKeyMatchesCurrentStateNaturalKeys(t *testing.T) {
	pair := base64.RawURLEncoding.EncodeToString([]byte("node1\x00/dev/sdb"))
	image := base64.RawURLEncoding.EncodeToString([]byte("rbd/image"))
	tests := []struct{ kind, path, want string }{
		{"device", "device/" + pair + "/zap", "node1:/dev/sdb"},
		{"rbd_image", "rbd/image/" + image, image},
		{"rbd_snapshot", "rbd/image/" + image + "/snapshot/snap1", image + "@snap1"},
		{"subvolume", "filesystem/cephfs/subvolume/data", "cephfs/data"},
		{"pool", "pool/data", "data"},
	}
	for _, test := range tests {
		if got := resourceLookupKey(test.kind, test.path); got != test.want {
			t.Errorf("resourceLookupKey(%q, %q) = %q, want %q", test.kind, test.path, got, test.want)
		}
	}
}

func TestEnqueueRejectsResourceChangedAfterPlanCreation(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	now := time.Now().UTC()
	name := "test"
	resource := store.CephResourceRecord{ClusterID: cluster.ID, Kind: "pool", NaturalKey: "test", Name: &name, Generation: 1, ResourceVersion: 7, Source: "fixture", ObservedAt: now, PayloadSchemaVersion: 1, PayloadJSON: `{}`, CreatedAt: now, UpdatedAt: now}
	if err := db.Insert(context.Background(), &resource); err != nil {
		t.Fatal(err)
	}
	plan, err := service.CreatePlan(context.Background(), PlanRequest{ClusterID: cluster.ID, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: 7, Risk: cephdomain.RiskHigh, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	resource.ResourceVersion = 8
	if err := db.UpsertResources(context.Background(), []store.CephResourceRecord{resource}); err != nil {
		t.Fatal(err)
	}
	generation := uint64(7)
	_, err = service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "pool.delete", ResourceKind: "pool", ResourceKey: "pool/test", ResourceGeneration: &generation, Risk: cephdomain.RiskHigh, PlanID: &plan.ID, Parameters: map[string]any{}})
	if err == nil || !strings.Contains(err.Error(), "generation changed") {
		t.Fatalf("stale plan enqueue error = %v", err)
	}
}
func TestWorkerCompletesAndAuditsOperation(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	if err := service.Register("test.success", func(context.Context, store.CephOperation) (cephdomain.OperationResult, error) {
		return cephdomain.OperationResult{Details: map[string]any{"ok": true}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	row, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, RequestID: "request", Action: "test.success", ResourceKind: "test", ResourceKey: "test/1", Risk: cephdomain.RiskLow, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	service.runOne(context.Background())
	row, err = service.Get(context.Background(), row.ID)
	if err != nil || row.Status != "succeeded" || row.Progress != 100 {
		t.Fatalf("operation=%#v err=%v", row, err)
	}
	audits, err := db.ListAuditEvents(context.Background(), cluster.ID, 20)
	if err != nil || len(audits) < 3 {
		t.Fatalf("audits=%d err=%v", len(audits), err)
	}
	for i := 0; i+1 < len(audits); i++ {
		if audits[i].PreviousHash == nil || *audits[i].PreviousHash != audits[i+1].EventHash {
			t.Fatalf("audit chain is broken between %d and %d", audits[i].ID, audits[i+1].ID)
		}
	}
	events, err := db.ListClusterOperationEventsAfter(context.Background(), cluster.ID, 0, 100)
	if err != nil || len(events) < 3 {
		t.Fatalf("stream events=%d err=%v", len(events), err)
	}
	lastID := events[len(events)-1].ID
	events, err = db.ListClusterOperationEventsAfter(context.Background(), cluster.ID, lastID, 100)
	if err != nil || len(events) != 0 {
		t.Fatalf("events after cursor=%d err=%v", len(events), err)
	}
}

func TestWorkerRequiresPostSuccessReconcileBeforeCompletion(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	if err := service.Register("test.reconcile", func(context.Context, store.CephOperation) (cephdomain.OperationResult, error) {
		return cephdomain.OperationResult{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	called := false
	service.SetPostSuccess(func(_ context.Context, operation store.CephOperation) error {
		called = operation.Action == "test.reconcile"
		return errors.New("refresh unavailable")
	})
	row, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "test.reconcile", ResourceKind: "test", ResourceKey: "test/1", Risk: cephdomain.RiskLow, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	service.runOne(context.Background())
	row, err = service.Get(context.Background(), row.ID)
	if err != nil || !called || row.Status != "needs_review" || row.ErrorCode == nil || *row.ErrorCode != "post_reconcile_failed" || row.Retryable {
		t.Fatalf("called=%v operation=%#v err=%v", called, row, err)
	}
}

func TestWorkerMarksUncertainPostCheckForReview(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	if err := service.Register("test.post-check", func(context.Context, store.CephOperation) (cephdomain.OperationResult, error) {
		return cephdomain.OperationResult{}, &cephdomain.OperationError{Code: "post_check_failed", Message: "state could not be confirmed", Retryable: true}
	}); err != nil {
		t.Fatal(err)
	}
	row, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "test.post-check", ResourceKind: "test", ResourceKey: "test/1", Risk: cephdomain.RiskMedium, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	service.runOne(context.Background())
	row, err = service.Get(context.Background(), row.ID)
	if err != nil || row.Status != "needs_review" || row.Retryable {
		t.Fatalf("operation=%#v err=%v", row, err)
	}
}

func TestWorkerRejectsGenerationChangedAfterEnqueue(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	now := time.Now().UTC()
	name := "data"
	resource := store.CephResourceRecord{ClusterID: cluster.ID, Kind: "pool", NaturalKey: "data", Name: &name, Generation: 1, ResourceVersion: 1, Source: "fixture", ObservedAt: now, PayloadSchemaVersion: 1, PayloadJSON: `{}`, CreatedAt: now, UpdatedAt: now}
	if err := db.Insert(context.Background(), &resource); err != nil {
		t.Fatal(err)
	}
	called := false
	if err := service.Register("pool.update", func(context.Context, store.CephOperation) (cephdomain.OperationResult, error) {
		called = true
		return cephdomain.OperationResult{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	generation := uint64(1)
	operation, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "pool.update", ResourceKind: "pool", ResourceKey: "pool/data", ResourceGeneration: &generation, Risk: cephdomain.RiskMedium, Parameters: map[string]any{"size": 3}})
	if err != nil {
		t.Fatal(err)
	}
	resource.ResourceVersion = 2
	if err := db.UpsertResources(context.Background(), []store.CephResourceRecord{resource}); err != nil {
		t.Fatal(err)
	}
	service.runOne(context.Background())
	operation, _ = service.Get(context.Background(), operation.ID)
	if called || operation.Status != "failed" || operation.ErrorCode == nil || *operation.ErrorCode != "resource_conflict" {
		t.Fatalf("called=%v operation=%#v", called, operation)
	}
}

func TestConcurrentAuditEventsRemainOneLinearChain(t *testing.T) {
	db, cluster := operationDatabase(t)
	const count = 8
	errCh := make(chan error, count)
	for i := 0; i < count; i++ {
		go func(index int) {
			now := time.Now().UTC().Add(time.Duration(index) * time.Nanosecond)
			errCh <- db.CreateAuditEvent(context.Background(), &store.AuditEvent{OccurredAt: now, EventType: "test", RequestID: fmt.Sprintf("request-%d", index), ActorUsername: "tester", ClusterID: &cluster.ID, Action: "audit.test", Outcome: "succeeded"})
		}(i)
	}
	for i := 0; i < count; i++ {
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	}
	rows, err := db.ListAuditEvents(context.Background(), cluster.ID, count)
	if err != nil || len(rows) != count {
		t.Fatalf("audits=%d err=%v", len(rows), err)
	}
	hashes := make(map[string]bool, len(rows))
	children := make(map[string]int, len(rows))
	for _, row := range rows {
		hashes[row.EventHash] = true
	}
	roots := 0
	for _, row := range rows {
		if row.PreviousHash == nil {
			roots++
			continue
		}
		if !hashes[*row.PreviousHash] {
			t.Fatalf("audit %d points outside the chain", row.ID)
		}
		children[*row.PreviousHash]++
		if children[*row.PreviousHash] > 1 {
			t.Fatalf("audit hash %s has multiple children", *row.PreviousHash)
		}
	}
	if roots != 1 {
		t.Fatalf("audit chain has %d roots", roots)
	}
}
func TestRunningOperationCanBeCancelled(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	started := make(chan struct{})
	_ = service.Register("test.block", func(ctx context.Context, _ store.CephOperation) (cephdomain.OperationResult, error) {
		close(started)
		<-ctx.Done()
		return cephdomain.OperationResult{}, ctx.Err()
	})
	row, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "test.block", ResourceKind: "test", ResourceKey: "test/1", Risk: cephdomain.RiskLow, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { service.runOne(context.Background()); close(done) }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not start")
	}
	if err := service.Cancel(context.Background(), row.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop")
	}
	row, _ = service.Get(context.Background(), row.ID)
	if row.Status != "cancelled" {
		t.Fatalf("status=%s", row.Status)
	}
}

func TestRecoveryRequeuesLowRiskAndReviewsHigherRisk(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 1)
	low, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "test.read", ResourceKind: "test", ResourceKey: "low", Risk: cephdomain.RiskLow, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	medium, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "test.write", ResourceKind: "test", ResourceKey: "medium", Risk: cephdomain.RiskMedium, Parameters: map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	stale := time.Now().UTC().Add(-time.Minute)
	for _, id := range []string{low.ID, medium.ID} {
		if err := db.UpdateOperation(context.Background(), id, map[string]any{"status": "running", "heartbeat_at": stale}); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.RecoverOperations(context.Background(), time.Now().UTC().Add(-30*time.Second), time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	low, _ = service.Get(context.Background(), low.ID)
	medium, _ = service.Get(context.Background(), medium.ID)
	if low.Status != "queued" || medium.Status != "needs_review" {
		t.Fatalf("unexpected recovery states: low=%s medium=%s", low.Status, medium.Status)
	}
}

func TestClaimLimitsConcurrentOperationsPerCluster(t *testing.T) {
	db, cluster := operationDatabase(t)
	service := New(func() *store.Database { return db }, 4)
	for i := 0; i < 3; i++ {
		if _, err := service.Enqueue(context.Background(), Request{ClusterID: &cluster.ID, ClusterName: cluster.Name, Action: "test.block", ResourceKind: "test", ResourceKey: fmt.Sprintf("item-%d", i), Risk: cephdomain.RiskLow, Parameters: map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err := db.ClaimQueuedOperation(context.Background(), time.Now().UTC(), 2); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ClaimQueuedOperation(context.Background(), time.Now().UTC(), 2); !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("third cluster operation was claimed: %v", err)
	}
}
