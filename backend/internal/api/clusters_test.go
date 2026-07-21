package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
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
		"description": "production ceph",
		"fsid": "00000000-0000-0000-0000-000000000001",
		"enabled": true,
		"dashboard": {
			"enabled": true,
			"base_url": "https://ceph.example.com:8443/",
			"username": "admin",
			"password": "dashboard-secret",
			"insecure_tls": true
		},
		"command": {
			"enabled": true,
			"bin": "ceph",
			"cluster": "ceph",
			"conf": "/etc/ceph/ceph.conf",
			"name": "client.admin",
			"keyring": "/etc/ceph/ceph.client.admin.keyring",
			"keyring_content": "[client.admin]\nkey = command-secret",
			"timeout_seconds": 30
		}
	}`)

	recorder := clusterAPIRequest(server, http.MethodPost, "/api/v1/clusters", adminToken, createPayload)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create cluster = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	var created cephClusterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created cluster: %v", err)
	}
	if !created.Dashboard.PasswordSet || !created.Command.KeyringContentSet {
		t.Fatalf("created = %#v, want masked secrets marked as set", created)
	}
	if created.Dashboard.BaseURL != "https://ceph.example.com:8443" {
		t.Fatalf("dashboard base_url = %q, want trimmed URL", created.Dashboard.BaseURL)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("dashboard-secret")) || bytes.Contains(recorder.Body.Bytes(), []byte("command-secret")) {
		t.Fatalf("response leaked cluster secrets: %s", recorder.Body.String())
	}

	updatePayload := []byte(`{
		"name": "primary-renamed",
		"enabled": true,
		"dashboard": {
			"enabled": true,
			"base_url": "https://ceph.example.com:8443",
			"username": "admin"
		},
		"command": {
			"enabled": true,
			"bin": "ceph",
			"timeout_seconds": 45
		}
	}`)
	recorder = clusterAPIRequest(server, http.MethodPut, "/api/v1/clusters/1", adminToken, updatePayload)
	if recorder.Code != http.StatusOK {
		t.Fatalf("update cluster = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var updated cephClusterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated cluster: %v", err)
	}
	if updated.Name != "primary-renamed" || !updated.Dashboard.PasswordSet || !updated.Command.KeyringContentSet || updated.Command.TimeoutSeconds != 45 {
		t.Fatalf("updated = %#v, want renamed cluster with retained secrets", updated)
	}

	clearPayload := []byte(`{
		"name": "primary-renamed",
		"enabled": true,
		"dashboard": {
			"enabled": true,
			"base_url": "https://ceph.example.com:8443",
			"username": "admin",
			"clear_secret": true
		},
		"command": {
			"enabled": true,
			"bin": "ceph",
			"clear_secret": true,
			"timeout_seconds": 45
		}
	}`)
	recorder = clusterAPIRequest(server, http.MethodPut, "/api/v1/clusters/1", adminToken, clearPayload)
	if recorder.Code != http.StatusOK {
		t.Fatalf("clear cluster secrets = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var cleared cephClusterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &cleared); err != nil {
		t.Fatalf("decode cleared cluster: %v", err)
	}
	if cleared.Dashboard.PasswordSet || cleared.Command.KeyringContentSet {
		t.Fatalf("cleared = %#v, want secret markers cleared", cleared)
	}

	recorder = clusterAPIRequest(server, http.MethodGet, "/api/v1/clusters", adminToken, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("list clusters = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var clusters []cephClusterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &clusters); err != nil {
		t.Fatalf("decode cluster list: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "primary-renamed" {
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

	return NewServer(cfg, nil, db), token
}

func clusterAPIRequest(server *Server, method, path, token string, body []byte) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	server.Routes().ServeHTTP(recorder, request)
	return recorder
}
