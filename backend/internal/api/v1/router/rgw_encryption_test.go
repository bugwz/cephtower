package router

import (
	"bytes"
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
	"cephtower/backend/internal/store"
)

type encryptionRouteExecutor struct {
	calls int
	fail  bool
}

func (e *encryptionRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls++
	if e.fail {
		return executor.CommandResult{ExitCode: 5, Stdout: []byte("private-error")}, nil
	}
	value := ""
	if strings.HasSuffix(spec.Args[3], "_backend") {
		value = "kmip"
	}
	if strings.HasSuffix(spec.Args[3], "_password") {
		value = "private-password"
	}
	return executor.CommandResult{Stdout: []byte(value + "\n")}, nil
}

func TestRGWEncryptionConfigurationAPI(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "encryption.db"}}, t.TempDir())
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
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, key, nil)
	runner := &encryptionRouteExecutor{}
	mux := http.NewServeMux()
	authEnabled := false
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, Inspection: clusterinspect.New(clusters, runner), AuthEnabled: func() bool { return authEnabled }}))
	send := func(body map[string]any) *httptest.ResponseRecorder {
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		r := httptest.NewRequest("GET", "/api/v1/rgw/encryption/configuration", bytes.NewReader(encoded))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	for _, body := range []map[string]any{{}, {"entity": "client.admin", "encryption_type": "kms", "provider": "vault"}, {"entity": "client.rgw.a", "encryption_type": "s3", "provider": "kmip"}, {"entity": 1, "encryption_type": "kms", "provider": "vault"}, {"entity": "client.rgw.a", "encryption_type": "kms", "provider": "vault", "unknown": true}} {
		if w := send(body); w.Code != 400 {
			t.Fatalf("invalid request: %d %s", w.Code, w.Body.String())
		}
	}
	if runner.calls != 0 {
		t.Fatal("invalid request issued commands")
	}
	body := map[string]any{"entity": "client.rgw.a", "encryption_type": "kms", "provider": "kmip"}
	w := send(body)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || strings.Contains(w.Body.String(), "private-") || !strings.Contains(w.Body.String(), "[REDACTED]") || runner.calls != 10 {
		t.Fatalf("read: %d %s", w.Code, w.Body.String())
	}
	runner.fail = true
	w = send(body)
	if w.Code == 200 || strings.Contains(w.Body.String(), "private-") {
		t.Fatalf("failure: %d %s", w.Code, w.Body.String())
	}
	authEnabled = true
	patchRequest := httptest.NewRequest("PATCH", "/api/v1/rgw/encryption/configuration", strings.NewReader(`{}`))
	patchRequest.Header.Set("Content-Type", "application/json")
	patchResponse := httptest.NewRecorder()
	mux.ServeHTTP(patchResponse, patchRequest)
	if patchResponse.Code != 401 {
		t.Fatal("unauthenticated encryption write accepted")
	}
	before := runner.calls
	if w := send(body); w.Code != 401 || runner.calls != before {
		t.Fatal("unauthenticated read reached native command")
	}
}
