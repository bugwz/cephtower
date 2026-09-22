package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
	"cephtower/backend/internal/service/reconciler"
	"cephtower/backend/internal/store"
)

type authRouteExecutor struct{ specs []executor.CommandSpec }

func (e *authRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	switch spec.ID {
	case "cluster.logs":
		return executor.CommandResult{Stdout: []byte(`[{"name":"mon.a","rank":"0","stamp":"2026-09-14 01:00:00","seq":1,"channel":"audit","priority":"[INF]","message":"entry"}]`)}, nil
	case "configuration.help":
		return executor.CommandResult{Stdout: []byte(`{"name":"osd_memory_target","type":"size","default":"4G","can_update_at_runtime":true}`)}, nil
	case "collect.ceph_user":
		return executor.CommandResult{Stdout: []byte(`{"auth_dump":[{"entity":"client.backup","key":"sensitive-fixture-key","caps":{"mon":"allow r"}}]}`)}, nil
	case "collect.config":
		return executor.CommandResult{Stdout: []byte(`[]`)}, nil
	case "ceph_user.export":
		return executor.CommandResult{Stdout: []byte("[client.backup]\n key = sensitive-fixture-key\n")}, nil
	default:
		return executor.CommandResult{}, nil
	}
}

func TestCephUserAPIEndToEndWithoutCluster(t *testing.T) {
	const encryptionKey = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: encryptionKey, SQLite: config.SQLiteConfig{Name: "auth-route.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	key, err := security.Encrypt([]byte("fixture"), encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: key, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	runner := &authRouteExecutor{}
	provider := &cephprovider.NativeProvider{Executor: runner}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, encryptionKey, provider)
	mutations := mutationservice.New(clusters, runner)
	reconcileService := reconciler.New(database, clusters, provider, reconciler.Options{})
	operations := operationservice.New(database, encryptionKey, operationservice.NewActionDispatcher(mutations, nil, reconcileService), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	h := handler.New(handler.Dependencies{Inspection: clusterinspect.New(clusters, runner), Database: database, Clusters: clusters, Mutations: mutations, Operations: operations, Reconciler: reconcileService, AuthEnabled: func() bool { return false }})
	mux := http.NewServeMux()
	Register(mux, h)
	send := func(method, path string, body map[string]any) *httptest.ResponseRecorder {
		t.Helper()
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(encoded))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK && rec.Code != http.StatusAccepted {
			t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
		if rec.Code == http.StatusAccepted {
			var response struct {
				Data struct {
					OperationID uint64 `json:"operation_id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Data.OperationID == 0 {
				t.Fatalf("decode accepted operation: id=%d err=%v", response.Data.OperationID, err)
			}
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				row, err := db.FindOperation(context.Background(), response.Data.OperationID)
				if err == nil && row.Status == store.OperationSucceeded {
					return rec
				}
				if err == nil && row.Status == store.OperationFailed {
					t.Fatalf("%s %s operation failed: %v %v", method, path, row.ErrorCode, row.ErrorMessage)
				}
				time.Sleep(time.Millisecond)
			}
			t.Fatalf("%s %s operation did not finish", method, path)
		}
		return rec
	}
	logResult := send("GET", "/logs", map[string]any{"channel": "audit", "level": "debug", "limit": 30})
	if !strings.Contains(logResult.Body.String(), "entry") {
		t.Fatal("log API did not return Ceph records")
	}
	helpResult := send("GET", "/configuration/option", map[string]any{"name": "osd_memory_target"})
	if !strings.Contains(helpResult.Body.String(), "can_update_at_runtime") {
		t.Fatal("configuration metadata missing")
	}
	send("PUT", "/configuration/value", map[string]any{"who": "osd/host:node-a", "name": "osd_memory_target", "value": "4G"})
	if _, err := db.FindResource(context.Background(), cluster.ID, "config_value", "osd/host:node-a:osd_memory_target"); !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("configuration cache must reflect the empty Ceph response: %v", err)
	}
	send("DELETE", "/configuration/value", map[string]any{"who": "osd/host:node-a", "name": "osd_memory_target"})
	send("PUT", "/configuration/value", map[string]any{"who": "client.rgw", "name": "rgw_keystone_admin_password", "value": "sensitive-config-value"})
	if _, err := db.FindResource(context.Background(), cluster.ID, "config_value", "client.rgw:rgw_keystone_admin_password"); !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatal("configuration request body entered the observed-state cache")
	}

	send("POST", "/resource/refresh", map[string]any{"kind": "ceph_user"})
	rec := send("GET", "/ceph/users", map[string]any{})
	if strings.Contains(rec.Body.String(), "sensitive-fixture-key") || !strings.Contains(rec.Body.String(), "allow r") {
		t.Fatal("list leaked secret or lost caps")
	}
	send("POST", "/ceph/user", map[string]any{"entity": "client.new", "caps": map[string]any{"mon": "allow r"}})
	send("PATCH", "/ceph/user", map[string]any{"entity": "client.backup", "caps": map[string]any{"osd": "profile rbd pool=data"}})
	rec = send("GET", "/ceph/users/export", map[string]any{"entities": []string{"client.backup"}})
	if rec.Header().Get("Cache-Control") != "no-store" || !strings.Contains(rec.Body.String(), "sensitive-fixture-key") {
		t.Fatal("export did not return a no-store keyring")
	}
	send("POST", "/ceph/users/import", map[string]any{"keyring": "[client.imported]\nkey = sensitive-import-key\n"})
	send("DELETE", "/ceph/user", map[string]any{"entity": "client.backup"})
	for _, spec := range runner.specs {
		if strings.Contains(fmt.Sprint(spec.Args), "sensitive-import-key") {
			t.Fatal("import key in argv")
		}
	}
	audits, err := db.ListAuditEvents(context.Background(), cluster.ID, 100)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, audit := range audits {
		if audit.ParametersJSON != nil && (strings.Contains(*audit.ParametersJSON, "sensitive-import-key") || strings.Contains(*audit.ParametersJSON, "sensitive-config-value")) {
			t.Fatal("import secret entered audit")
		}
		if audit.Action == "ceph_user.import" {
			found = true
		}
	}
	if !found {
		t.Fatal("import action not audited")
	}
	h.AuthEnabled = func() bool { return true }
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/ceph/users/export", strings.NewReader(`{"cluster_id":1,"entities":["client.backup"]}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated export: %d", rec.Code)
	}
}
