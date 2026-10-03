package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/service/clusterinspect"
	mutationservice "cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type globalScheduleRunner struct{}

func (globalScheduleRunner) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.ID == "rbd_mirroring.global_schedule.post_check" || spec.ID == "rbd.mirror.schedules" {
		return executor.CommandResult{Stdout: []byte(`[{"pool":"-","namespace":"-","image":"-","items":[{"interval":"1h","start_time":""}]}]`)}, nil
	}
	return executor.CommandResult{}, nil
}

func TestGlobalMirrorScheduleAPI(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "global-schedule.db"}}, t.TempDir())
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
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "rbd", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, key, nil)
	mutations := mutationservice.New(clusters, globalScheduleRunner{})
	operations := operationservice.New(database, key, operationservice.NewActionDispatcher(mutations, nil, nil), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, Inspection: clusterinspect.New(clusters, globalScheduleRunner{}), Mutations: mutations, Operations: operations, AuthEnabled: func() bool { return false }}))
	query, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID})
	read := httptest.NewRecorder()
	mux.ServeHTTP(read, httptest.NewRequest("GET", "/api/v1/rbd/mirroring/schedules", strings.NewReader(string(query))))
	if read.Code != http.StatusOK || read.Header().Get("Cache-Control") != "no-store" || !strings.Contains(read.Body.String(), `"schedules":[{"pool":"-"`) {
		t.Fatal(read.Code, read.Body.String())
	}
	body, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID, "action": "mirror-schedule-add", "interval": "1h"})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/rbd/mirroring/global/schedule", strings.NewReader(string(body))))
	if rec.Code != http.StatusAccepted {
		t.Fatal(rec.Code, rec.Body.String())
	}
	var response struct {
		Data struct {
			OperationID uint64 `json:"operation_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		row, err := db.FindOperation(context.Background(), response.Data.OperationID)
		if err == nil && row.Status == store.OperationSucceeded {
			if row.ResourceKey != "global/snapshot-schedule" || row.Risk != "high" {
				t.Fatal(row)
			}
			return
		}
		if err == nil && row.Status == store.OperationFailed {
			data, _ := json.Marshal(row)
			t.Fatalf("operation failed: %s", data)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("global schedule operation did not finish")
}
