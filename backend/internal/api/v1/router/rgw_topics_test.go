package router

import (
	"context"
	"encoding/json"
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
	"cephtower/backend/internal/service/reconciler"
	"cephtower/backend/internal/store"
)

type topicAPIExecutor struct {
	filesystemFieldsExecutor
	unavailable bool
}

func (e *topicAPIExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	switch spec.ID {
	case "collect.rgw_topic":
		if e.unavailable {
			return executor.CommandResult{}, fmt.Errorf("topic metadata unavailable")
		}
		return executor.CommandResult{Stdout: []byte(`["tenant:events"]`)}, nil
	case "collect.rgw_topic_detail":
		return executor.CommandResult{Stdout: []byte(`{"key":"topic:tenant:events","ver":{"tag":"t","ver":9007199254740993},"data":{"name":"events","owner":"tenant$user","arn":"arn:aws:sns:zone:tenant:events","dest":{"push_endpoint":"https://user:secret-value@host/path?token=secret-value","push_endpoint_args":"password=secret-value","persistent":false,"max_retries":"18446744073709551615"}}}`)}, nil
	}
	return e.filesystemFieldsExecutor.Run(ctx, access, spec)
}

func TestRGWTopicsNativeToAPI(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "topics.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	secret, _ := security.Encrypt([]byte("fixture"), key)
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "topics", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: secret, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "rgw_admin", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	exec := &topicAPIExecutor{}
	provider := &cephprovider.NativeProvider{Executor: exec}
	clusters := clusterservice.New(database, key, provider)
	service := reconciler.New(database, clusters, provider, reconciler.Options{})
	if _, err := service.RefreshKinds(context.Background(), cluster.ID, []string{"rgw_topic"}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, AuthEnabled: func() bool { return false }}))
	for _, unavailable := range []bool{false, true} {
		if unavailable {
			exec.unavailable = true
			_, _ = service.RefreshKinds(context.Background(), cluster.ID, []string{"rgw_topic"})
		}
		body, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID})
		req := httptest.NewRequest("GET", "/api/v1/rgw/topics", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var response struct {
			Data struct {
				Items []struct {
					Data map[string]any `json:"data"`
				} `json:"items"`
			} `json:"data"`
		}
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &response) != nil || len(response.Data.Items) != 1 {
			t.Fatalf("%d %s", rec.Code, rec.Body.String())
		}
		data := response.Data.Items[0].Data
		if data["metadata_version"] != `{"tag":"t","ver":9007199254740993}` {
			t.Fatalf("native version rounded: %v", data["metadata_version"])
		}
		if data["name"] != "events" || data["scope"] != "tenant" || data["push_endpoint"] != "https://host/path" || data["persistent"] != false || data["max_retries"] != "18446744073709551615" || strings.Contains(rec.Body.String(), "secret-value") {
			t.Fatalf("bad topic response %s", rec.Body.String())
		}
	}
}
