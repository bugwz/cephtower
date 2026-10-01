package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	mutationservice "cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type filesystemRenameRouteExecutor struct{ specs []executor.CommandSpec }

func (e *filesystemRenameRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.ID == "filesystem.rename.post_check" {
		return executor.CommandResult{Stdout: []byte(`[{"name":"renamed"}]`)}, nil
	}
	return executor.CommandResult{Stdout: []byte("volume renamed")}, nil
}

func TestFilesystemRenameRouteWithoutCluster(t *testing.T) {
	const encryptionKey = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: encryptionKey, SQLite: config.SQLiteConfig{Name: "rename-route.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	key, _ := security.Encrypt([]byte("fixture"), encryptionKey)
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: key, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "cephfs_volume", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	runner := &filesystemRenameRouteExecutor{}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, encryptionKey, &cephprovider.NativeProvider{Executor: runner})
	mutations := mutationservice.New(clusters, runner)
	operations := operationservice.New(database, encryptionKey, operationservice.NewActionDispatcher(mutations, nil, nil), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, Mutations: mutations, Operations: operations, AuthEnabled: func() bool { return false }}))
	send := func(body map[string]any) *httptest.ResponseRecorder {
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		req := httptest.NewRequest("PUT", "/api/v1/filesystem", bytes.NewReader(encoded))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	for _, body := range []map[string]any{
		{"fs": "original", "new_name": "renamed"},
		{"fs": "original", "new_name": "renamed", "confirmed": "true"},
		{"fs": "original", "new_name": "renamed", "confirmed": true, "unexpected": true},
	} {
		if rec := send(body); rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid contract accepted: %d %s", rec.Code, rec.Body.String())
		}
	}
	rec := send(map[string]any{"fs": "original", "new_name": "renamed", "confirmed": true})
	if rec.Code != http.StatusAccepted {
		t.Fatalf("rename route: %d %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			OperationID uint64 `json:"operation_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Data.OperationID == 0 {
		t.Fatalf("invalid operation response: %s", rec.Body.String())
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		row, err := db.FindOperation(context.Background(), response.Data.OperationID)
		if err == nil && row.Status == store.OperationSucceeded {
			if row.Risk != "high" || row.Action != "filesystem.rename" || row.ResourceKey != "filesystem/original" || len(runner.specs) != 2 || !reflect.DeepEqual(runner.specs[0].Args, []string{"fs", "volume", "rename", "original", "renamed", "--yes-i-really-mean-it"}) {
				t.Fatalf("unexpected operation: %+v, specs: %+v", row, runner.specs)
			}
			return
		}
		if err == nil && row.Status == store.OperationFailed {
			t.Fatalf("rename failed: %+v", row)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("rename operation did not finish")
}
