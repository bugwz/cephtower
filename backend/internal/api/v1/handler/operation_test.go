package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
	"cephtower/backend/internal/config"
	clusterservice "cephtower/backend/internal/service/cluster"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

func TestMutationQueuesInspectableOperation(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: contractKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "operations-api.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, contractKey, unusedProvider{})
	operations := operationservice.New(database, contractKey, nil, operationservice.Options{})
	h := handler.New(handler.Dependencies{Clusters: clusters, Operations: operations, Database: database, AuthEnabled: func() bool { return false }})
	mux := http.NewServeMux()
	router.Register(mux, h)

	queued := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/pool", fmt.Sprintf(`{"cluster_id":%d,"name":"data"}`, cluster.ID), "create-data")
	if queued.Code != http.StatusAccepted {
		t.Fatalf("queue status=%d body=%s", queued.Code, queued.Body.String())
	}
	operationID := operationIDFromResponse(t, queued)
	row, err := db.FindOperation(context.Background(), operationID)
	if err != nil || row.Status != store.OperationQueued || strings.Contains(row.ParametersCiphertext, `"name":"data"`) {
		t.Fatalf("queued operation=%#v err=%v", row, err)
	}

	replayed := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/pool", fmt.Sprintf(`{"cluster_id":%d,"name":"data"}`, cluster.ID), "create-data")
	if replayed.Code != http.StatusAccepted || operationIDFromResponse(t, replayed) != operationID {
		t.Fatalf("idempotent replay status=%d body=%s", replayed.Code, replayed.Body.String())
	}
	conflict := sendOperationRequest(t, mux, http.MethodDelete, "/api/v1/pool", fmt.Sprintf(`{"cluster_id":%d,"name":"data"}`, cluster.ID), "create-data")
	if conflict.Code != http.StatusConflict {
		t.Fatalf("idempotency conflict status=%d body=%s", conflict.Code, conflict.Body.String())
	}

	get := sendOperationRequest(t, mux, http.MethodGet, "/api/v1/operation", fmt.Sprintf(`{"cluster_id":%d,"operation_id":%d}`, cluster.ID, operationID), "")
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), "parameters_ciphertext") {
		t.Fatalf("get operation status=%d body=%s", get.Code, get.Body.String())
	}
	list := sendOperationRequest(t, mux, http.MethodGet, "/api/v1/operations?status=queued&limit=10", fmt.Sprintf(`{"cluster_id":%d}`, cluster.ID), "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"operation_id":`) {
		t.Fatalf("list operations status=%d body=%s", list.Code, list.Body.String())
	}
}

func sendOperationRequest(t *testing.T, mux http.Handler, method, path, body, idempotencyKey string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}

func operationIDFromResponse(t *testing.T, response *httptest.ResponseRecorder) uint64 {
	t.Helper()
	var envelope struct {
		Data struct {
			OperationID uint64 `json:"operation_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope.Data.OperationID == 0 {
		t.Fatalf("decode operation response: id=%d err=%v body=%s", envelope.Data.OperationID, err, response.Body.String())
	}
	return envelope.Data.OperationID
}
