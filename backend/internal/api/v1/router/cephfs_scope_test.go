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
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	mutationservice "cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type scopedCephFSExecutor struct{ specs []executor.CommandSpec }

func (e *scopedCephFSExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	return executor.CommandResult{Stdout: []byte(`{"status":{"state":"canceled"}}`)}, nil
}

func TestCephFSGroupScopedReadsAndVersionedOperations(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "cephfs-scope.db"}}, t.TempDir())
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
	rows := []store.CephEntityRecord{}
	for _, group := range []string{"_nogroup", "team-a", "team-b"} {
		parent := "cephfs/" + group + "/same"
		fs, parentKind, name, snapshotKind := "cephfs", "filesystem", "same", "subvolume"
		payload, _ := json.Marshal(map[string]any{"fs": fs, "group": group, "name": name})
		rows = append(rows, store.CephEntityRecord{Kind: "subvolume", NaturalKey: parent, ParentKind: &parentKind, ParentKey: &fs, Name: &name, Source: "native", DiscoveredData: string(payload), ObservedAt: now})
		snapshotName := "daily"
		payload, _ = json.Marshal(map[string]any{"fs": fs, "group": group, "subvolume": name, "name": snapshotName})
		rows = append(rows, store.CephEntityRecord{Kind: "cephfs_snapshot", NaturalKey: parent + "/daily", ParentKind: &snapshotKind, ParentKey: &parent, Name: &snapshotName, Source: "native", DiscoveredData: string(payload), ObservedAt: now})
	}
	if err := db.ReconcileResources(context.Background(), cluster.ID, 1, rows, []string{"subvolume", "cephfs_snapshot"}); err != nil {
		t.Fatal(err)
	}
	// Only team-b has version 2: selecting another group's version must fail.
	for i := range rows {
		if i >= 4 {
			rows[i].DiscoveredData = rows[i].DiscoveredData[:len(rows[i].DiscoveredData)-1] + `,"changed":true}`
		}
	}
	if err := db.ReconcileResources(context.Background(), cluster.ID, 2, rows, []string{"subvolume", "cephfs_snapshot"}); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	runner := &scopedCephFSExecutor{}
	clusters := clusterservice.New(database, key, nil)
	mutations := mutationservice.New(clusters, runner)
	operations := operationservice.New(database, key, operationservice.NewActionDispatcher(mutations, nil, nil), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, Mutations: mutations, Operations: operations, AuthEnabled: func() bool { return false }}))
	send := func(method, path, version string, body map[string]any) *httptest.ResponseRecorder {
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(encoded))
		req.Header.Set("Content-Type", "application/json")
		if version != "" {
			req.Header.Set("If-Match", version)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	for _, group := range []string{"", "_nogroup", "team-a", "team-b"} {
		wantGroup := group
		if wantGroup == "" {
			wantGroup = "_nogroup"
		}
		body := map[string]any{"fs": "cephfs", "subvolume": "same", "group": group}
		rec := send("GET", "/filesystem/subvolume", "", body)
		var result struct {
			Data struct {
				NaturalKey      string `json:"natural_key"`
				ResourceVersion uint64 `json:"resource_version"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || rec.Code != 200 || result.Data.NaturalKey != "cephfs/"+wantGroup+"/same" {
			t.Fatalf("read group %s: %d %s", group, rec.Code, rec.Body.String())
		}
		wantVersion := uint64(1)
		if group == "team-b" {
			wantVersion = 2
		}
		if result.Data.ResourceVersion != wantVersion {
			t.Fatal(result)
		}
	}
	for _, tt := range []struct {
		body  map[string]any
		count int
	}{
		{map[string]any{}, 3}, {map[string]any{"fs": "cephfs"}, 3},
		{map[string]any{"fs": "cephfs", "subvolume": "same", "group": "team-b"}, 1},
		{map[string]any{"fs": "cephfs", "subvolume": "same"}, 1},
		{map[string]any{"group": "team-a"}, 1},
	} {
		rec := send("GET", "/filesystem/subvolume/snapshots", "", tt.body)
		var result struct {
			Data struct {
				Items []map[string]any `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || rec.Code != 200 || len(result.Data.Items) != tt.count {
			t.Fatalf("snapshot scope %v: %d %s", tt.body, rec.Code, rec.Body.String())
		}
		if tt.count == 1 {
			wantGroup, _ := tt.body["group"].(string)
			if wantGroup == "" {
				wantGroup = "_nogroup"
			}
			data, ok := result.Data.Items[0]["data"].(map[string]any)
			if !ok || data["group"] != wantGroup {
				t.Fatalf("wrong group returned: %s", rec.Body.String())
			}
		}
	}
	for _, tt := range []struct {
		method, path string
		body         map[string]any
	}{
		{"PATCH", "/filesystem/subvolume", map[string]any{"fs": "cephfs", "subvolume": "same", "group": "team-b", "size": 2048}},
		{"DELETE", "/filesystem/subvolume/snapshot", map[string]any{"fs": "cephfs", "subvolume": "same", "group": "team-b", "snap": "daily"}},
	} {
		if rec := send(tt.method, tt.path, "1", tt.body); rec.Code != 409 {
			t.Fatalf("wrong group version accepted: %d %s", rec.Code, rec.Body.String())
		}
	}
	if len(runner.specs) != 0 {
		t.Fatal("conflicted requests executed commands")
	}
	for _, tt := range []struct {
		method, path, version, group, lock string
		body                               map[string]any
		args                               []string
	}{
		{"PATCH", "/filesystem/subvolume", "2", "team-b", "cephfs/team-b/same", map[string]any{"size": 2048}, []string{"fs", "subvolume", "resize", "cephfs", "same", "2048", "team-b"}},
		{"PATCH", "/filesystem/subvolume", "1", "team-a", "cephfs/team-a/same", map[string]any{"size": 2048}, []string{"fs", "subvolume", "resize", "cephfs", "same", "2048", "team-a"}},
		{"PATCH", "/filesystem/subvolume", "1", "", "cephfs/_nogroup/same", map[string]any{"size": 2048}, []string{"fs", "subvolume", "resize", "cephfs", "same", "2048"}},
		{"DELETE", "/filesystem/subvolume/snapshot", "2", "team-b", "cephfs/team-b/same/daily", map[string]any{"snap": "daily"}, []string{"fs", "subvolume", "snapshot", "rm", "cephfs", "same", "daily", "team-b"}},
		{"POST", "/filesystem/subvolume/clone/cancel", "2", "team-b", "cephfs/team-b/same", map[string]any{}, []string{"fs", "clone", "cancel", "cephfs", "same", "--group_name", "team-b"}},
	} {
		tt.body["fs"], tt.body["subvolume"], tt.body["group"] = "cephfs", "same", tt.group
		before := len(runner.specs)
		rec := send(tt.method, tt.path, tt.version, tt.body)
		if rec.Code != 202 {
			t.Fatalf("operation: %d %s", rec.Code, rec.Body.String())
		}
		var result struct {
			Data struct {
				OperationID uint64 `json:"operation_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(2 * time.Second)
		for {
			row, err := db.FindOperation(context.Background(), result.Data.OperationID)
			if err == nil && row.Status == store.OperationSucceeded {
				if row.LockKey != tt.lock || len(runner.specs) != before+2 || !reflect.DeepEqual(runner.specs[before].Args, tt.args) {
					t.Fatalf("operation=%+v specs=%+v", row, runner.specs[before:])
				}
				break
			}
			if err == nil && row.Status == store.OperationFailed {
				t.Fatalf("operation failed=%+v", row)
			}
			if time.Now().After(deadline) {
				t.Fatal("operation did not finish")
			}
			time.Sleep(time.Millisecond)
		}
	}
}
