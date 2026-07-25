package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	clusterservice "cephtower/backend/internal/service/cluster"
	service "cephtower/backend/internal/service/collector"
	"cephtower/backend/internal/store"
)

func TestClusterRoutesManageCephConnections(t *testing.T) {
	h, db, admin := newClusterAPITestAPI(t, store.UserRoleAdmin, nil)
	defer func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	createPayload := []byte(`{
		"name": "primary",
		"monitor_host": "10.0.0.11:6789,10.0.0.12:6789",
		"keyring": "command-secret",
		"dashboard_username": "admin",
		"dashboard_password": "dashboard-secret"
	}`)

	recorder := clusterAPIRequest(h.CreateCluster, http.MethodPost, "/api/v1/cluster", admin, createPayload, nil)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create cluster = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	var createResponse MessageResponse
	if err := decodeAPIResponseData(recorder, &createResponse); err != nil {
		t.Fatalf("decode created cluster: %v", err)
	}
	if createResponse.Message == "" {
		t.Fatalf("create response = %#v, want message", createResponse)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("dashboard-secret")) || bytes.Contains(recorder.Body.Bytes(), []byte("command-secret")) {
		t.Fatalf("response leaked cluster secrets: %s", recorder.Body.String())
	}

	recorder = clusterAPIRequest(h.ListClusters, http.MethodGet, "/api/v1/cluster", admin, nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("list clusters after create = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var clusters []CephClusterResponse
	if err := decodeAPIResponseData(recorder, &clusters); err != nil {
		t.Fatalf("decode cluster list after create: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "primary" || clusters[0].Dashboard.Password != "dashboard-secret" || clusters[0].Command.Keyring != "command-secret" {
		t.Fatalf("clusters = %#v, want created cluster with plaintext secrets", clusters)
	}
	if clusters[0].Command.Bin != "ceph" || clusters[0].Command.Name != "client.admin" || clusters[0].Command.TimeoutSeconds != 15 {
		t.Fatalf("cluster command = %#v, want default ceph client.admin command", clusters[0].Command)
	}

	updatePayload := []byte(`{
		"name": "primary-renamed",
		"monitor_host": "",
		"keyring": "",
		"dashboard_username": "admin",
		"dashboard_password": ""
	}`)
	recorder = clusterAPIRequest(h.UpdateCluster, http.MethodPut, "/api/v1/cluster/1", admin, updatePayload, map[string]string{"id": "1"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("update cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var updateResponse MessageResponse
	if err := decodeAPIResponseData(recorder, &updateResponse); err != nil {
		t.Fatalf("decode updated cluster: %v", err)
	}
	if updateResponse.Message == "" {
		t.Fatalf("update response = %#v, want message", updateResponse)
	}

	recorder = clusterAPIRequest(h.ListClusters, http.MethodGet, "/api/v1/cluster", admin, nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("list clusters = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if err := decodeAPIResponseData(recorder, &clusters); err != nil {
		t.Fatalf("decode cluster list: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "primary-renamed" || clusters[0].Dashboard.Password != "dashboard-secret" || clusters[0].Command.Keyring != "command-secret" {
		t.Fatalf("clusters = %#v, want updated cluster in list", clusters)
	}
}

func TestClusterRoutesRequireAdministrator(t *testing.T) {
	h, db, user := newClusterAPITestAPI(t, store.UserRoleUser, nil)
	defer func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	recorder := clusterAPIRequest(h.ListClusters, http.MethodGet, "/api/v1/cluster", user, nil, nil)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("list clusters as user = %d, want 403", recorder.Code)
	}
}

func TestDeleteClusterRemovesDiscoveredResources(t *testing.T) {
	h, db, admin := newClusterAPITestAPI(t, store.UserRoleAdmin, nil)
	defer func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	cluster := store.CephCluster{
		Name:              "delete-me",
		MonitorHost:       "10.0.0.11:6789",
		Keyring:           "command-secret",
		DashboardUsername: "admin",
		DashboardPassword: "dashboard-secret",
	}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	if err := db.Insert(context.Background(), &store.CephClusterHost{
		ClusterID:    cluster.ID,
		Hostname:     "node-1",
		Labels:       `[]`,
		Sources:      `{}`,
		Payload:      `{"hostname":"node-1"}`,
		DiscoveredAt: time.Now(),
	}); err != nil {
		t.Fatalf("create discovered host: %v", err)
	}
	if err := db.Insert(context.Background(), &store.CephClusterMon{
		ClusterID:    cluster.ID,
		Name:         "a",
		Payload:      `{"name":"a"}`,
		DiscoveredAt: time.Now(),
	}); err != nil {
		t.Fatalf("create discovered mon: %v", err)
	}

	recorder := clusterAPIRequest(h.DeleteCluster, http.MethodDelete, "/api/v1/cluster/1", admin, nil, map[string]string{"id": "1"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("delete cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var response MessageResponse
	if err := decodeAPIResponseData(recorder, &response); err != nil {
		t.Fatalf("decode delete response: %v", err)
	}
	if response.Message == "" {
		t.Fatalf("delete response = %#v, want message", response)
	}

	assertModelCount(t, db, &store.CephCluster{}, 0)
	assertModelCount(t, db, &store.CephClusterHost{}, 0)
	assertModelCount(t, db, &store.CephClusterMon{}, 0)
	assertModelCount(t, db, &store.CephDataFetchRun{}, 0)
}

func TestCreateClusterStoresDiscoveredCephInventory(t *testing.T) {
	discoverer := func(ctx context.Context, db *store.Database, cluster *store.CephCluster) error {
		return db.Insert(ctx, &store.CephClusterHost{
			ClusterID:    cluster.ID,
			Hostname:     "node-1",
			Addr:         "10.0.0.1",
			Labels:       `[]`,
			Sources:      `{}`,
			Payload:      `{"hostname":"node-1","addr":"10.0.0.1"}`,
			DiscoveredAt: time.Now(),
		})
	}
	h, db, admin := newClusterAPITestAPI(t, store.UserRoleAdmin, discoverer)
	defer func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	payload := []byte(`{
		"name": "discovered",
		"monitor_host": "10.0.0.11:6789",
		"keyring": "command-secret",
		"dashboard_username": "admin",
		"dashboard_password": "dashboard-secret"
	}`)

	recorder := clusterAPIRequest(h.CreateCluster, http.MethodPost, "/api/v1/cluster", admin, payload, nil)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create discovered cluster = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	var createResponse MessageResponse
	if err := decodeAPIResponseData(recorder, &createResponse); err != nil {
		t.Fatalf("decode created cluster: %v", err)
	}
	if createResponse.Message == "" {
		t.Fatalf("create response = %#v, want message", createResponse)
	}

	var created store.CephCluster
	if err := db.FindRecord(context.Background(), map[string]any{"name": "discovered"}, &created); err != nil {
		t.Fatalf("load created cluster: %v", err)
	}
	if created.MonitorHost != "10.0.0.11:6789" || created.Keyring != "command-secret" || created.DashboardUsername != "admin" || created.DashboardPassword != "dashboard-secret" {
		t.Fatalf("created = %#v, want submitted cluster connection fields", created)
	}

	var host store.CephClusterHost
	if err := db.FindClusterRecord(context.Background(), created.ID, map[string]any{"hostname": "node-1"}, &host); err != nil {
		t.Fatalf("load discovered host: %v", err)
	}
	if !bytes.Contains([]byte(host.Payload), []byte("node-1")) {
		t.Fatalf("host payload = %s, want discovered host", host.Payload)
	}

	settings, err := db.ListSettings(context.Background(), service.DataFetchSettingPrefix)
	if err != nil {
		t.Fatalf("count system data fetch settings: %v", err)
	}
	if len(settings) == 0 {
		t.Fatal("system data fetch settings are empty; want defaults from cluster creation")
	}

	recorder = clusterAPIRequest(h.GetCluster, http.MethodGet, "/api/v1/cluster/1", admin, nil, map[string]string{"id": "1"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("get discovered cluster detail = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var detail CephClusterDetailResponse
	if err := decodeAPIResponseData(recorder, &detail); err != nil {
		t.Fatalf("decode cluster detail: %v", err)
	}
	if detail.Cluster.ID != created.ID || len(detail.Discovery.Hosts) != 1 {
		t.Fatalf("detail = %#v, want cluster and discovered inventory", detail)
	}
}

func assertModelCount(t *testing.T, db *store.Database, model any, want int64) {
	t.Helper()

	got, err := db.CountRecords(context.Background(), model, nil)
	if err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if got != want {
		t.Fatalf("%T count = %d, want %d", model, got, want)
	}
}

func newClusterAPITestAPI(t *testing.T, role string, discoverer service.ClusterDiscoverer) (*Handler, *store.Database, store.User) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "cephtower.db")
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Engine: store.EngineSQLite,
			SQLite: config.SQLiteConfig{Path: dbPath},
		},
	}
	db, err := store.Open(cfg.Database)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}

	passwordHash, err := store.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}
	user := store.User{
		Username:     "tester",
		DisplayName:  "Tester",
		Role:         role,
		Permissions:  `["cluster:read","storage:read","system:read"]`,
		PasswordHash: passwordHash,
		Enabled:      true,
	}
	if err := db.CreateUser(context.Background(), &user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if discoverer == nil {
		discoverer = func(_ context.Context, _ *store.Database, _ *store.CephCluster) error {
			return nil
		}
	}

	h := New(nil, Dependencies{ClusterService: clusterservice.New(clusterservice.Dependencies{
		Database: func() *store.Database { return db }, Discover: discoverer,
	})})
	return h, db, user
}

func clusterAPIRequest(h http.HandlerFunc, method, path string, user store.User, body []byte, pathValues map[string]string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	for key, value := range pathValues {
		request.SetPathValue(key, value)
	}
	request = request.WithContext(ContextWithUser(request.Context(), user))
	h(recorder, request)
	return recorder
}
