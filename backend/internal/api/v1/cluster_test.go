package v1

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
	"cephtower/backend/internal/service/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func TestClusterRoutesManageCephConnections(t *testing.T) {
	api, db, admin := newClusterAPITestAPI(t, store.UserRoleAdmin, nil)
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

	recorder := clusterAPIRequest(api.CreateCluster, http.MethodPost, "/api/v1/cluster", admin, createPayload, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("create cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var createResponse clusterActionResponse
	if err := decodeAPIResponseData(recorder, &createResponse); err != nil {
		t.Fatalf("decode created cluster: %v", err)
	}
	if createResponse.Message == "" {
		t.Fatalf("create response = %#v, want message", createResponse)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("dashboard-secret")) || bytes.Contains(recorder.Body.Bytes(), []byte("command-secret")) {
		t.Fatalf("response leaked cluster secrets: %s", recorder.Body.String())
	}

	recorder = clusterAPIRequest(api.ListClusters, http.MethodGet, "/api/v1/cluster", admin, nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("list clusters after create = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var clusters []cephClusterResponse
	if err := decodeAPIResponseData(recorder, &clusters); err != nil {
		t.Fatalf("decode cluster list after create: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "primary" || !clusters[0].Dashboard.PasswordSet || !clusters[0].Command.KeyringContentSet {
		t.Fatalf("clusters = %#v, want created cluster with masked secrets marked as set", clusters)
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
	recorder = clusterAPIRequest(api.UpdateCluster, http.MethodPut, "/api/v1/cluster/1", admin, updatePayload, map[string]string{"id": "1"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("update cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var updateResponse clusterActionResponse
	if err := decodeAPIResponseData(recorder, &updateResponse); err != nil {
		t.Fatalf("decode updated cluster: %v", err)
	}
	if updateResponse.Message == "" {
		t.Fatalf("update response = %#v, want message", updateResponse)
	}

	recorder = clusterAPIRequest(api.ListClusters, http.MethodGet, "/api/v1/cluster", admin, nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("list clusters = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if err := decodeAPIResponseData(recorder, &clusters); err != nil {
		t.Fatalf("decode cluster list: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "primary-renamed" || !clusters[0].Dashboard.PasswordSet || !clusters[0].Command.KeyringContentSet {
		t.Fatalf("clusters = %#v, want updated cluster in list", clusters)
	}
}

func TestClusterRoutesRequireAdministrator(t *testing.T) {
	api, db, user := newClusterAPITestAPI(t, store.UserRoleUser, nil)
	defer func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	recorder := clusterAPIRequest(api.ListClusters, http.MethodGet, "/api/v1/cluster", user, nil, nil)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("list clusters as user = %d, want 403", recorder.Code)
	}
}

func TestDeleteClusterRemovesDiscoveredResources(t *testing.T) {
	api, db, admin := newClusterAPITestAPI(t, store.UserRoleAdmin, nil)
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
	if err := db.Create(&cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	if err := db.Create(&store.CephClusterHost{
		ClusterID:    cluster.ID,
		Hostname:     "node-1",
		Labels:       `[]`,
		Sources:      `{}`,
		Payload:      `{"hostname":"node-1"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create discovered host: %v", err)
	}
	if err := db.Create(&store.CephClusterMon{
		ClusterID:    cluster.ID,
		Name:         "a",
		Payload:      `{"name":"a"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create discovered mon: %v", err)
	}

	recorder := clusterAPIRequest(api.DeleteCluster, http.MethodDelete, "/api/v1/cluster/1", admin, nil, map[string]string{"id": "1"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("delete cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var response clusterActionResponse
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
	discoverer := func(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
		return db.WithContext(ctx).Create(&store.CephClusterHost{
			ClusterID:    cluster.ID,
			Hostname:     "node-1",
			Addr:         "10.0.0.1",
			Labels:       `[]`,
			Sources:      `{}`,
			Payload:      `{"hostname":"node-1","addr":"10.0.0.1"}`,
			DiscoveredAt: time.Now(),
		}).Error
	}
	api, db, admin := newClusterAPITestAPI(t, store.UserRoleAdmin, discoverer)
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

	recorder := clusterAPIRequest(api.CreateCluster, http.MethodPost, "/api/v1/cluster", admin, payload, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("create discovered cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var createResponse clusterActionResponse
	if err := decodeAPIResponseData(recorder, &createResponse); err != nil {
		t.Fatalf("decode created cluster: %v", err)
	}
	if createResponse.Message == "" {
		t.Fatalf("create response = %#v, want message", createResponse)
	}

	var created store.CephCluster
	if err := db.Where("name = ?", "discovered").First(&created).Error; err != nil {
		t.Fatalf("load created cluster: %v", err)
	}
	if created.MonitorHost != "10.0.0.11:6789" || created.Keyring != "command-secret" || created.DashboardUsername != "admin" || created.DashboardPassword != "dashboard-secret" {
		t.Fatalf("created = %#v, want submitted cluster connection fields", created)
	}

	var host store.CephClusterHost
	if err := db.
		Where("cluster_id = ? AND hostname = ?", created.ID, "node-1").
		First(&host).Error; err != nil {
		t.Fatalf("load discovered host: %v", err)
	}
	if !bytes.Contains([]byte(host.Payload), []byte("node-1")) {
		t.Fatalf("host payload = %s, want discovered host", host.Payload)
	}

	var settingCount int64
	if err := db.Model(&store.Setting{}).Where("`key` LIKE ?", ceph.DataFetchSettingPrefix+"%").Count(&settingCount).Error; err != nil {
		t.Fatalf("count system data fetch settings: %v", err)
	}
	if settingCount == 0 {
		t.Fatalf("system data fetch settings = %d, want defaults from cluster creation", settingCount)
	}

	recorder = clusterAPIRequest(api.GetCluster, http.MethodGet, "/api/v1/cluster/1", admin, nil, map[string]string{"id": "1"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("get discovered cluster detail = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var detail cephClusterDetailResponse
	if err := decodeAPIResponseData(recorder, &detail); err != nil {
		t.Fatalf("decode cluster detail: %v", err)
	}
	if detail.Cluster.ID != created.ID || len(detail.Discovery.Hosts) != 1 {
		t.Fatalf("detail = %#v, want cluster and discovered inventory", detail)
	}
}

func assertModelCount(t *testing.T, db *gorm.DB, model any, want int64) {
	t.Helper()

	var got int64
	if err := db.Model(model).Count(&got).Error; err != nil {
		t.Fatalf("count %T: %v", model, err)
	}
	if got != want {
		t.Fatalf("%T count = %d, want %d", model, got, want)
	}
}

func newClusterAPITestAPI(t *testing.T, role string, discoverer ceph.ClusterDiscoverer) (*API, *gorm.DB, store.User) {
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
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if discoverer == nil {
		discoverer = func(_ context.Context, _ *gorm.DB, _ *store.CephCluster) error {
			return nil
		}
	}

	api := NewAPI(nil, Dependencies{
		CurrentConfig: func() config.Config {
			return cfg
		},
		Database: func() *gorm.DB {
			return db
		},
		ClusterDiscoverer: discoverer,
	})
	return api, db, user
}

func clusterAPIRequest(handler http.HandlerFunc, method, path string, user store.User, body []byte, pathValues map[string]string) *httptest.ResponseRecorder {
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
	handler(recorder, request)
	return recorder
}
