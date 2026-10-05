package router

import (
	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/service/clusterinspect"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"cephtower/backend/internal/service/external"
	"cephtower/backend/internal/store"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type daemonRouteExecutor struct {
	calls   int
	output  string
	exit    int
	outputs map[string]string
}

func (e *daemonRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls++
	if output, ok := e.outputs[spec.ID]; ok {
		return executor.CommandResult{Stdout: []byte(output), ExitCode: e.exit}, nil
	}
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
	Register(mux, handler.New(handler.Dependencies{Database: database, External: external.New(endpointservice.New(database, key), key, nil), Clusters: clusterservice.New(database, key, nil), Inspection: clusterinspect.New(clusterservice.New(database, key, nil), runner), AuthEnabled: func() bool { return auth }}))
	path := "/api/v1/rgw/daemons"
	send := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, strings.NewReader(body))
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
	path = "/api/v1/rgw/daemon/status"
	runner.exit = 0
	runner.output = `{"rgw":{"1":{"status_stamp":"stamp","last_beacon":"beacon","status":{"current_sync":"idle","password":"private-secret","json":"{\"count\":9007199254740993}"}}}}`
	statusBody := strings.TrimSuffix(valid, "}") + `,"service_map_id":"1"}`
	w = send(statusBody)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || strings.Contains(w.Body.String(), "private-secret") || !strings.Contains(w.Body.String(), `9007199254740993`) {
		t.Fatal("unsafe status", w.Code)
	}
	var envelope struct {
		Data struct {
			Status map[string]string `json:"status"`
		} `json:"data"`
	}
	if json.Unmarshal(w.Body.Bytes(), &envelope) != nil || envelope.Data.Status["json"] != `{"count":9007199254740993}` {
		t.Fatal("nested json lost string precision")
	}
	count := runner.calls
	for _, body := range []string{valid, strings.TrimSuffix(valid, "}") + `,"service_map_id":1}`, strings.TrimSuffix(statusBody, "}") + `,"name":"unexpected"}`} {
		if w = send(body); w.Code != 400 {
			t.Fatal("invalid status request accepted")
		}
	}
	if runner.calls != count {
		t.Fatal("invalid status request ran command")
	}
	runner.output = `{"rgw":{}}`
	if w = send(statusBody); w.Code != 404 {
		t.Fatal("missing status not reported", w.Code)
	}
	runner.exit = 3
	runner.output = "private-failure"
	if w = send(statusBody); w.Code == 200 || strings.Contains(w.Body.String(), "private-") {
		t.Fatal("status failure leaked")
	}
	path = "/api/v1/rgw/daemon/perf"
	for _, body := range []string{`{}`, valid, `{"cluster_id":0,"service_map_id":"1"}`, strings.TrimSuffix(valid, "}") + `,"service_map_id":1}`, strings.TrimSuffix(statusBody, "}") + `,"query":"up"}`} {
		if w = send(body); w.Code != 400 {
			t.Fatal("invalid performance request accepted", w.Code)
		}
	}
	if w = send(statusBody); w.Code != 501 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing monitoring endpoint not reported", w.Code)
	}
	path = "/api/v1/rgw/daemon/perf/history"
	if w = send(statusBody); w.Code != 501 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("history endpoint unavailable not reported")
	}
	if w = send(strings.TrimSuffix(statusBody, "}") + `,"start":"arbitrary"}`); w.Code != 400 {
		t.Fatal("custom history window accepted")
	}
	path = "/api/v1/rgw/topology/counts"
	runner.exit = 0
	runner.output = `{"realms":[],"zonegroups":["g"],"zones":["z"]}`
	if w = send(valid); w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), `"realm_count":0`) || !strings.Contains(w.Body.String(), `"zone_count":1`) {
		t.Fatal("invalid topology count response", w.Code)
	}
	count = runner.calls
	if w = send(`{"cluster_id":0}`); w.Code != 400 || runner.calls != count {
		t.Fatal("invalid count scope executed")
	}
	if w = send(strings.TrimSuffix(valid, "}") + `,"realm":"other"}`); w.Code != 400 || runner.calls != count {
		t.Fatal("unexpected count scope accepted")
	}
	path = "/api/v1/rgw/buckets/usage"
	runner.outputs = map[string]string{"rgw.usage.list": `["photos"]`, "rgw.usage.stats": `{"bucket":"photos","tenant":"","usage":{"rgw.main":{"num_objects":9007199254740993,"size_actual":18446744073709551615}}}`}
	w = send(valid)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), `"object_count":"9007199254740993"`) || !strings.Contains(w.Body.String(), `"size_actual_bytes":"18446744073709551615"`) {
		t.Fatal("usage precision lost", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"scope":"current_rgw_configuration"`) {
		t.Fatal("usage configuration scope missing")
	}
	count = runner.calls
	for _, body := range []string{`{}`, `{"cluster_id":0}`, `{"cluster_id":"1"}`, strings.TrimSuffix(valid, "}") + `,"bucket":"photos"}`} {
		if w = send(body); w.Code != 400 || runner.calls != count {
			t.Fatal("invalid usage scope executed")
		}
	}
	runner.outputs["rgw.usage.list"] = `[]`
	if w = send(valid); w.Code != 200 || !strings.Contains(w.Body.String(), `"object_count":"0"`) {
		t.Fatal("empty aggregate lost")
	}
	runner.outputs["rgw.usage.list"] = `["photos"]`
	runner.outputs["rgw.usage.stats"] = `{"bucket":"photos","tenant":""}`
	if w = send(valid); w.Code != 502 || strings.Contains(w.Body.String(), `"bucket_count"`) {
		t.Fatal("partial aggregate returned")
	}
	runner.exit = 2
	if w = send(valid); w.Code != 502 {
		t.Fatal("failed usage read accepted")
	}
	runner.exit = 0
	path = "/api/v1/rgw/users/count"
	runner.output = `["alice","tenant$alice"]`
	if w = send(valid); w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), `"user_count":2`) || !strings.Contains(w.Body.String(), `"scope":"current_rgw_configuration"`) || strings.Contains(w.Body.String(), "alice") {
		t.Fatal("invalid user count response", w.Code)
	}
	count = runner.calls
	for _, body := range []string{`{}`, `{"cluster_id":0}`, `{"cluster_id":"1"}`, strings.TrimSuffix(valid, "}") + `,"realm":"other"}`} {
		if w = send(body); w.Code != 400 || runner.calls != count {
			t.Fatal("invalid user count scope executed")
		}
	}
	runner.output = `[]`
	if w = send(valid); w.Code != 200 || !strings.Contains(w.Body.String(), `"user_count":0`) {
		t.Fatal("empty user count lost")
	}
	runner.output = `["a","a"]`
	if w = send(valid); w.Code != 502 || strings.Contains(w.Body.String(), `"user_count"`) {
		t.Fatal("invalid user count accepted")
	}
	runner.output = `[]`
	runner.exit = 2
	if w = send(valid); w.Code != 502 {
		t.Fatal("failed user list accepted")
	}
	runner.exit = 0
	auth = true
	count = runner.calls
	if w = send(valid); w.Code != 401 || runner.calls != count {
		t.Fatal("user count authentication bypass")
	}
	path = "/api/v1/rgw/buckets/usage"
	count = runner.calls
	if w = send(valid); w.Code != 401 || runner.calls != count {
		t.Fatal("usage authentication bypass")
	}
	path = "/api/v1/rgw/topology/counts"
	if w = send(valid); w.Code != 401 || runner.calls != count {
		t.Fatal("counts authentication bypass")
	}
	path = "/api/v1/rgw/daemon/perf/history"
	if w = send(statusBody); w.Code != 401 {
		t.Fatal("history authentication bypass")
	}
	path = "/api/v1/rgw/daemon/perf"
	if w = send(statusBody); w.Code != 401 {
		t.Fatal("performance authentication bypass")
	}
	path = "/api/v1/rgw/daemon/status"
	before := runner.calls
	if w = send(statusBody); w.Code != 401 || runner.calls != before {
		t.Fatal("authentication bypass")
	}
	path = "/api/v1/rgw/daemons"
	if w = send(valid); w.Code != 401 || runner.calls != before {
		t.Fatal("list authentication bypass")
	}
}
