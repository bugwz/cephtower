package handler_test

import (
	"context"
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
	"cephtower/backend/internal/store"
)

func TestResourceHistoryAPIOnlyExposesRetainedKinds(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: contractKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "history-api.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	history := store.CephObservationHistory{
		ClusterID: cluster.ID, Kind: "overview", NaturalKey: "overview", Source: "ceph_cli",
		ObservedAt: now, DataJSON: `{"health_status":"HEALTH_OK"}`, CreatedAt: now,
	}
	if _, err := db.AppendObservationHistoryIfDue(context.Background(), &history, 0); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, contractKey, unusedProvider{})
	mux := http.NewServeMux()
	router.Register(mux, handler.New(handler.Dependencies{Clusters: clusters, Database: database, AuthEnabled: func() bool { return false }}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/resource/history?limit=10", strings.NewReader(fmt.Sprintf(`{"cluster_id":%d,"kind":"overview","natural_key":"overview"}`, cluster.ID)))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"health_status":"HEALTH_OK"`) || strings.Contains(response.Body.String(), "data_json") {
		t.Fatalf("history status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/resource/history", strings.NewReader(fmt.Sprintf(`{"cluster_id":%d,"kind":"pool"}`, cluster.ID)))
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "history_not_supported") {
		t.Fatalf("unsupported history status=%d body=%s", response.Code, response.Body.String())
	}
}
