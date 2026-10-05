package router

import (
	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/service/clusterinspect"
	"cephtower/backend/internal/store"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type daemonRouteExecutor struct {
	calls  int
	output string
	exit   int
}

func (e *daemonRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, _ executor.CommandSpec) (executor.CommandResult, error) {
	e.calls++
	return executor.CommandResult{Stdout: []byte(e.output), ExitCode: e.exit}, nil
}

func TestRGWDaemonAPI(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "daemon.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	secret, _ := security.Encrypt([]byte("fixture"), key)
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: secret, CreatedAt: now, UpdatedAt: now}
	if err = db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	runner := &daemonRouteExecutor{output: `{"services":{"rgw":{"daemons":{"summary":"","1":{"metadata":{"id":"a","frontend_config#0":"private-secret"}}}}}}`}
	auth := false
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusterservice.New(database, key, nil), Inspection: clusterinspect.New(clusterservice.New(database, key, nil), runner), AuthEnabled: func() bool { return auth }}))
	send := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/api/v1/rgw/daemons", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	valid := fmt.Sprintf(`{"cluster_id":%d}`, cluster.ID)
	for _, body := range []string{`{}`, `{"cluster_id":0}`, `{"cluster_id":"1"}`, strings.TrimSuffix(valid, "}") + `,"name":"unexpected"}`} {
		if w := send(body); w.Code != 400 {
			t.Fatal("invalid request accepted", w.Code)
		}
	}
	if runner.calls != 0 {
		t.Fatal("invalid request ran command")
	}
	w := send(valid)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || strings.Contains(w.Body.String(), "private-") || !strings.Contains(w.Body.String(), `"service_map_id":"1"`) {
		t.Fatal("unsafe response", w.Code)
	}
	runner.output = `{"services":{}}`
	w = send(valid)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Fatal("empty result lost")
	}
	runner.output = "private-error"
	runner.exit = 2
	w = send(valid)
	if w.Code == 200 || strings.Contains(w.Body.String(), "private-error") {
		t.Fatal("command failure leaked")
	}
	auth = true
	before := runner.calls
	if w = send(valid); w.Code != 401 || runner.calls != before {
		t.Fatal("authentication bypass")
	}
}
