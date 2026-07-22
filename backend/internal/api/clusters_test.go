package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func TestClusterRoutesManageCephConnections(t *testing.T) {
	server, adminToken := newClusterAPITestServer(t, store.UserRoleAdmin)
	defer func() {
		if err := server.Close(); err != nil {
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

	recorder := clusterAPIRequest(server, http.MethodPost, "/api/v1/clusters", adminToken, createPayload)
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

	recorder = clusterAPIRequest(server, http.MethodGet, "/api/v1/clusters", adminToken, nil)
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
	recorder = clusterAPIRequest(server, http.MethodPut, "/api/v1/clusters/1", adminToken, updatePayload)
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

	recorder = clusterAPIRequest(server, http.MethodGet, "/api/v1/clusters", adminToken, nil)
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
	server, userToken := newClusterAPITestServer(t, store.UserRoleUser)
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	recorder := clusterAPIRequest(server, http.MethodGet, "/api/v1/clusters", userToken, nil)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("list clusters as user = %d, want 403", recorder.Code)
	}
}

func TestDeleteClusterRemovesDiscoveredResources(t *testing.T) {
	server, adminToken := newClusterAPITestServer(t, store.UserRoleAdmin)
	defer func() {
		if err := server.Close(); err != nil {
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
	if err := server.database().Create(&cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	if err := server.database().Create(&store.CephClusterHost{
		ClusterID:    cluster.ID,
		Hostname:     "node-1",
		Labels:       `[]`,
		Sources:      `{}`,
		Payload:      `{"hostname":"node-1"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create discovered host: %v", err)
	}
	if err := server.database().Create(&store.CephClusterMon{
		ClusterID:    cluster.ID,
		Name:         "a",
		Payload:      `{"name":"a"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create discovered mon: %v", err)
	}

	recorder := clusterAPIRequest(server, http.MethodDelete, "/api/v1/clusters/1", adminToken, nil)
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

	assertModelCount(t, server.database(), &store.CephCluster{}, 0)
	assertModelCount(t, server.database(), &store.CephClusterHost{}, 0)
	assertModelCount(t, server.database(), &store.CephClusterMon{}, 0)
	assertModelCount(t, server.database(), &store.CephDataFetchRun{}, 0)
}

func TestCreateClusterStoresDiscoveredCephInventory(t *testing.T) {
	server, adminToken := newClusterAPITestServer(t, store.UserRoleAdmin)
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()
	server.clusterDiscoverer = func(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
		saveDiscoveredHosts(ctx, db, cluster.ID, func() ([]ceph.Host, error) {
			return []ceph.Host{
				{Hostname: "node-1", Addr: "10.0.0.1"},
			}, nil
		})
		return nil
	}

	payload := []byte(`{
		"name": "discovered",
		"monitor_host": "10.0.0.11:6789",
		"keyring": "command-secret",
		"dashboard_username": "admin",
		"dashboard_password": "dashboard-secret"
	}`)

	recorder := clusterAPIRequest(server, http.MethodPost, "/api/v1/clusters", adminToken, payload)
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
	if err := server.database().Where("name = ?", "discovered").First(&created).Error; err != nil {
		t.Fatalf("load created cluster: %v", err)
	}
	if created.MonitorHost != "10.0.0.11:6789" || created.Keyring != "command-secret" || created.DashboardUsername != "admin" || created.DashboardPassword != "dashboard-secret" {
		t.Fatalf("created = %#v, want submitted cluster connection fields", created)
	}

	var host store.CephClusterHost
	if err := server.database().
		Where("cluster_id = ? AND hostname = ?", created.ID, "node-1").
		First(&host).Error; err != nil {
		t.Fatalf("load discovered host: %v", err)
	}
	if !bytes.Contains([]byte(host.Payload), []byte("node-1")) {
		t.Fatalf("host payload = %s, want discovered host", host.Payload)
	}

	var settingCount int64
	if err := server.database().Model(&store.Setting{}).Where("`key` LIKE ?", dataFetchSettingPrefix+"%").Count(&settingCount).Error; err != nil {
		t.Fatalf("count system data fetch settings: %v", err)
	}
	if settingCount == 0 {
		t.Fatalf("system data fetch settings = %d, want defaults from cluster creation", settingCount)
	}

	recorder = clusterAPIRequest(server, http.MethodGet, "/api/v1/clusters/1", adminToken, nil)
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

func newClusterAPITestServer(t *testing.T, role string) (*Server, string) {
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
	token := "test-token"
	session := store.UserSession{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}

	server := NewServer(cfg, nil, db)
	server.clusterDiscoverer = func(_ context.Context, _ *gorm.DB, _ *store.CephCluster) error {
		return nil
	}
	return server, token
}

func clusterAPIRequest(server *Server, method, path, token string, body []byte) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	server.Routes().ServeHTTP(recorder, request)
	return recorder
}

func decodeAPIResponseData(recorder *httptest.ResponseRecorder, out any) error {
	response, err := decodeAPIResponseEnvelope(recorder)
	if err != nil {
		return err
	}
	if response.Code != 0 {
		return errors.New(response.Message)
	}
	return json.Unmarshal(response.Data, out)
}

type apiResponseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeAPIResponseEnvelope(recorder *httptest.ResponseRecorder) (apiResponseEnvelope, error) {
	var response apiResponseEnvelope
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	return response, err
}
