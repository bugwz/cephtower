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
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "rgw_admin", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	for index, tc := range []struct{ fields, risk string }{
		{`"email":"before@example.test"`, "medium"},
		{`"account_root":true,"expected_account_id":"RGW12345678901234567"`, "high"},
		{`"account_root":false,"expected_account_id":"RGW12345678901234567"`, "high"},
		{`"target_account_id":"RGW12345678901234567","migration_confirm_uid":"tenant$user"`, "high"},
		{`"email":"after@example.test"`, "medium"},
	} {
		response := sendOperationRequest(t, mux, http.MethodPatch, "/api/v1/rgw/user", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user",%s}`, cluster.ID, tc.fields), fmt.Sprintf("user-risk-%d", index))
		if response.Code != http.StatusAccepted {
			t.Fatalf("user update: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != tc.risk || row.Action != "rgw_user.update" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("incorrect user risk or identity: %+v %v", row, err)
		}
		if !strings.Contains(response.Body.String(), `"risk":"`+tc.risk+`"`) {
			t.Fatalf("response risk differs: %s", response.Body.String())
		}
	}
	for _, action := range []string{"modify", "rm"} {
		fields := ""
		if action == "modify" {
			fields = `,"subuser_permission":"read"`
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user/subuser", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","subuser":"swift","confirm_subuser":"tenant$user:swift","action":%q%s}`, cluster.ID, action, fields), "subuser-"+action)
		if response.Code != http.StatusAccepted {
			t.Fatalf("subuser mutation: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_user.subuser" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("wrong subuser operation: %+v %v", row, err)
		}
	}
	for _, root := range []bool{false, true} {
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user", fmt.Sprintf(`{"cluster_id":%d,"uid":"new-user","display_name":"valid-name","account_id":"RGW12345678901234567","account_root":%t}`, cluster.ID, root), fmt.Sprintf("create-account-user-%t", root))
		if response.Code != http.StatusAccepted {
			t.Fatalf("account user create: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		want := "medium"
		if root {
			want = "high"
		}
		if err != nil || row.Risk != want {
			t.Fatalf("wrong creation risk: %+v %v", row, err)
		}
	}
	for _, kind := range []string{"s3", "swift"} {
		fields := ""
		if kind == "s3" {
			fields = `,"access_key":"ACCESS123"`
		}
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/rgw/user/subuser", fmt.Sprintf(`{"cluster_id":%d,"uid":"tenant$user","subuser":"new","confirm_subuser":"tenant$user:new","action":"create","subuser_permission":"read","key_type":%q,"secret_key":"PrivateTestSecret"%s}`, cluster.ID, kind, fields), "subuser-create-"+kind)
		if response.Code != http.StatusAccepted {
			t.Fatalf("subuser create: %d %s", response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "rgw_user.subuser" || row.ResourceKey != "rgw/user/tenant$user" {
			t.Fatalf("wrong creation operation: %+v %v", row, err)
		}
		if strings.Contains(row.ParametersCiphertext, "PrivateTestSecret") || strings.Contains(response.Body.String(), "PrivateTestSecret") || strings.Contains(response.Body.String(), "ACCESS123") {
			t.Fatal("creation credential exposed")
		}
	}
	for _, action := range []string{"start", "stop", "restart", "redeploy", "reconfig", "rotate-key"} {
		response := sendOperationRequest(t, mux, http.MethodPost, "/api/v1/daemon/action", fmt.Sprintf(`{"cluster_id":%d,"name":"osd.1","action":%q}`, cluster.ID, action), "daemon-"+action)
		if response.Code != http.StatusAccepted {
			t.Fatalf("daemon action %s: %d %s", action, response.Code, response.Body.String())
		}
		row, err := db.FindOperation(context.Background(), operationIDFromResponse(t, response))
		if err != nil || row.Risk != "high" || row.Action != "daemon.action" || row.ResourceKey != "daemon/osd.1/action" || row.Status != store.OperationQueued {
			t.Fatalf("incorrect daemon action risk or target: %+v %v", row, err)
		}
	}

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
