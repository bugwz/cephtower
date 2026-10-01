package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	mutationservice "cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type groupUpdateRouteExecutor struct {
	specs []executor.CommandSpec
	quota string
}

func (e *groupUpdateRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if len(spec.Args) >= 6 && spec.Args[2] == "resize" && spec.Args[5] != "inf" {
		e.quota = spec.Args[5]
	}
	quota := e.quota
	if quota == "" {
		quota = `"infinite"`
	}
	return executor.CommandResult{Stdout: []byte(fmt.Sprintf(`{"data_pool":"hot","uid":1000,"gid":1001,"mode":16832,"bytes_quota":%s}`, quota))}, nil
}

func TestSubvolumeGroupUpdateAPIWithoutCluster(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "group-update.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	secret, _ := security.Encrypt([]byte("fixture"), key)
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: secret, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "cephfs_volume", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	if err := db.ReconcileResources(context.Background(), cluster.ID, 1, []store.CephEntityRecord{{Kind: "subvolume_group", NaturalKey: "cephfs/users", Source: "native", DiscoveredData: "{}", ObservedAt: now}}, []string{"subvolume_group"}); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	runner := &groupUpdateRouteExecutor{}
	clusters := clusterservice.New(database, key, nil)
	mutations := mutationservice.New(clusters, runner)
	operations := operationservice.New(database, key, operationservice.NewActionDispatcher(mutations, nil, nil), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, Mutations: mutations, Operations: operations, AuthEnabled: func() bool { return false }}))
	send := func(body map[string]any) *httptest.ResponseRecorder {
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		req := httptest.NewRequest("PATCH", "/api/v1/filesystem/subvolume/group", bytes.NewReader(encoded))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("If-Match", "1")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	for _, body := range []map[string]any{
		{"group": "users", "unlimited": true}, {"fs": "cephfs", "unlimited": true},
		{"fs": "cephfs", "group": "users", "unlimited": "true"},
		{"fs": "cephfs", "group": "users", "size": 1.5},
		{"fs": "cephfs", "group": "users", "size": 1024},
		{"fs": "cephfs", "group": "users", "uid": "1000"},
		{"fs": "cephfs", "group": "users", "size": 1024, "extra": true},
	} {
		if rec := send(body); rec.Code != 400 {
			t.Fatalf("invalid request accepted: %d %s", rec.Code, rec.Body.String())
		}
	}
	rec := send(map[string]any{"fs": "cephfs", "group": "users", "unlimited": true, "pool": "hot", "uid": 1000, "gid": 1001, "mode": "0700"})
	if rec.Code != 202 {
		t.Fatalf("PATCH: %d %s", rec.Code, rec.Body.String())
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
			want := [][]string{
				{"fs", "subvolumegroup", "info", "cephfs", "users", "--format", "json"},
				{"fs", "subvolumegroup", "resize", "cephfs", "users", "inf"},
				{"fs", "subvolumegroup", "create", "cephfs", "users", "--pool_layout", "hot", "--uid", "1000", "--gid", "1001", "--mode", "0700"},
				{"fs", "subvolumegroup", "info", "cephfs", "users", "--format", "json"},
			}
			if row.Action != "subvolume_group.update" || row.Risk != "medium" || row.LockKey != "cephfs/users" || len(runner.specs) != 4 {
				t.Fatalf("operation=%+v specs=%+v", row, runner.specs)
			}
			for i, spec := range runner.specs {
				if !reflect.DeepEqual(spec.Args, want[i]) || spec.Mutating != (i == 1 || i == 2) {
					t.Fatalf("spec %d=%+v", i, spec)
				}
			}
			runner.specs = nil
			rec = send(map[string]any{"fs": "cephfs", "group": "users", "size": "9007199254740993", "no_shrink": true})
			if rec.Code != 202 {
				t.Fatalf("exact quota request: %d %s", rec.Code, rec.Body.String())
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			deadline = time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				operation, err := db.FindOperation(context.Background(), response.Data.OperationID)
				if err == nil && operation.Status == store.OperationSucceeded {
					if len(runner.specs) != 2 || !reflect.DeepEqual(runner.specs[0].Args, []string{"fs", "subvolumegroup", "resize", "cephfs", "users", "9007199254740993", "--no_shrink"}) {
						t.Fatalf("exact quota specs=%+v", runner.specs)
					}
					return
				}
				if err == nil && operation.Status == store.OperationFailed {
					t.Fatalf("exact quota operation=%+v", operation)
				}
				time.Sleep(time.Millisecond)
			}
			t.Fatal("exact quota operation timed out")
		}
		if err == nil && row.Status == store.OperationFailed {
			t.Fatalf("operation failed=%+v", row)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("group update operation did not finish")
}
