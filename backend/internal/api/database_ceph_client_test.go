package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func TestDatabaseCephClientUsesEnabledDashboardCluster(t *testing.T) {
	var authCalls int
	cephServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			authCalls++
			writeTestJSON(w, http.StatusCreated, map[string]any{"token": "test-token"})
		case "/api/summary":
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Fatalf("Authorization = %q, want bearer token", got)
			}
			writeTestJSON(w, http.StatusOK, map[string]any{
				"health_status": "HEALTH_OK",
				"version":       "20.2.2",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer cephServer.Close()
	addFakeCephCommand(t, map[string]any{"dashboard": cephServer.URL})

	db := openDatabaseCephClientTestDB(t)
	if err := db.Create(&store.CephCluster{
		Name:              "primary",
		MonitorHost:       "10.0.0.11:6789",
		Keyring:           "secret",
		DashboardUsername: "admin",
		DashboardPassword: "password",
	}).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	client := newDatabaseCephClient(func() *gorm.DB { return db })
	summary, err := client.ClusterSummary(context.Background())
	if err != nil {
		t.Fatalf("ClusterSummary() returned error: %v", err)
	}
	if summary.HealthStatus != "HEALTH_OK" || summary.Version != "20.2.2" {
		t.Fatalf("summary = %#v, want dashboard response", summary)
	}
	if authCalls != 1 {
		t.Fatalf("auth calls = %d, want 1", authCalls)
	}
}

func TestDatabaseCephClientListsHostsFromSnapshot(t *testing.T) {
	db := openDatabaseCephClientTestDB(t)
	cluster := store.CephCluster{
		Name:              "primary",
		MonitorHost:       "10.0.0.11:6789",
		Keyring:           "secret",
		DashboardUsername: "admin",
		DashboardPassword: "password",
	}
	if err := db.Create(&cluster).Error; err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	if err := db.Create(&store.CephResourceSnapshot{
		ClusterID:    cluster.ID,
		Category:     snapshotHosts,
		ResourceKey:  "all",
		Payload:      `[{"hostname":"node-a","addr":"10.0.0.1"},{"hostname":"node-b","addr":"10.0.0.2"}]`,
		LastSyncedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	client := newDatabaseCephClient(func() *gorm.DB { return db })
	hosts, err := client.ListHosts(context.Background(), ceph.ListHostsOptions{Search: "node-b"})
	if err != nil {
		t.Fatalf("ListHosts() returned error: %v", err)
	}
	if len(hosts) != 1 || hosts[0].Hostname != "node-b" {
		t.Fatalf("hosts = %#v, want node-b from snapshot", hosts)
	}
}

func TestDatabaseCephClientReturnsUnknownSummaryWithoutDashboardCluster(t *testing.T) {
	db := openDatabaseCephClientTestDB(t)
	client := newDatabaseCephClient(func() *gorm.DB { return db })

	summary, err := client.ClusterSummary(context.Background())
	if err != nil {
		t.Fatalf("ClusterSummary() returned error: %v", err)
	}
	if summary.HealthStatus != "unknown" {
		t.Fatalf("HealthStatus = %q, want unknown", summary.HealthStatus)
	}
}

func addFakeCephCommand(t *testing.T, payload any) {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal fake ceph payload: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "ceph")
	script := "#!/bin/sh\nprintf '%s\\n' '" + string(data) + "'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake ceph command: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func openDatabaseCephClientTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg := config.DatabaseConfig{
		Engine: store.EngineSQLite,
		SQLite: config.SQLiteConfig{Path: filepath.Join(t.TempDir(), "cephtower.db")},
	}
	db, err := store.Open(cfg)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	})
	return db
}

func writeTestJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
