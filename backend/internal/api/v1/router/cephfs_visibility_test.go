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
	"cephtower/backend/internal/service/clusterinspect"
	mutationservice "cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type visibilityRouteExecutor struct{ specs []executor.CommandSpec }

func (e *visibilityRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	return executor.CommandResult{Stdout: []byte("0\n")}, nil
}

func TestSubvolumeSnapshotVisibilityAPIWithoutCluster(t *testing.T) {
	const encryptionKey = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: encryptionKey, SQLite: config.SQLiteConfig{Name: "visibility.db"}}, t.TempDir())
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
	if err := db.ReconcileResources(context.Background(), cluster.ID, 1, []store.CephEntityRecord{{Kind: "subvolume", NaturalKey: "cephfs/users/home", Source: "native", DiscoveredData: "{}", ObservedAt: now}}, []string{"subvolume"}); err != nil {
		t.Fatal(err)
	}
	runner := &visibilityRouteExecutor{}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, encryptionKey, &cephprovider.NativeProvider{Executor: runner})
	mutations := mutationservice.New(clusters, runner)
	operations := operationservice.New(database, encryptionKey, operationservice.NewActionDispatcher(mutations, nil, nil), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, Mutations: mutations, Inspection: clusterinspect.New(clusters, runner), Operations: operations, AuthEnabled: func() bool { return false }}))
	send := func(method string, body map[string]any) *httptest.ResponseRecorder {
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api/v1/filesystem/subvolume/snapshot/visibility", bytes.NewReader(encoded))
		req.Header.Set("Content-Type", "application/json")
		if method == "PUT" {
			req.Header.Set("If-Match", "1")
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	for _, body := range []map[string]any{
		{"subvolume": "home", "visible": false},
		{"fs": "cephfs", "visible": false},
		{"fs": "cephfs", "subvolume": "home"},
		{"fs": "cephfs", "subvolume": "home", "visible": "false"},
		{"fs": "cephfs", "subvolume": "home", "visible": false, "unexpected": true},
	} {
		if rec := send("PUT", body); rec.Code != 400 {
			t.Fatalf("invalid request accepted: %d %s", rec.Code, rec.Body.String())
		}
	}
	rec := send("GET", map[string]any{"fs": "cephfs", "subvolume": "home", "group": "users"})
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("GET: %d %s", rec.Code, rec.Body.String())
	}
	var read struct {
		Data struct {
			Visible *bool  `json:"visible"`
			Group   string `json:"group"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &read); err != nil || read.Data.Visible == nil || *read.Data.Visible || read.Data.Group != "users" {
		t.Fatalf("GET response: %s", rec.Body.String())
	}
	rec = send("PUT", map[string]any{"fs": "cephfs", "subvolume": "home", "group": "users", "visible": false})
	if rec.Code != 202 {
		t.Fatalf("PUT: %d %s", rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			OperationID uint64 `json:"operation_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Data.OperationID == 0 {
		t.Fatal(rec.Body.String())
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		row, err := db.FindOperation(context.Background(), response.Data.OperationID)
		if err == nil && row.Status == store.OperationSucceeded {
			want := []string{"fs", "subvolume", "snapshot_visibility", "set", "cephfs", "home", "false", "--group_name", "users"}
			if row.Action != "subvolume.snapshot_visibility" || row.Risk != "medium" || row.ResourceKey != "filesystem/cephfs/subvolume/home/group/users" || row.LockKey != "cephfs/users/home" || len(runner.specs) != 3 || !reflect.DeepEqual(runner.specs[1].Args, want) {
				t.Fatalf("operation=%+v specs=%+v", row, runner.specs)
			}
			return
		}
		if err == nil && row.Status == store.OperationFailed {
			t.Fatalf("operation failed=%+v", row)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("visibility operation did not finish")
}
