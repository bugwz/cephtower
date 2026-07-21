package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func TestDatabaseCephClientUsesEnabledDashboardCluster(t *testing.T) {
	var authCalls int
	cephServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			authCalls++
			writeJSON(w, http.StatusCreated, map[string]any{"token": "test-token"})
		case "/api/summary":
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Fatalf("Authorization = %q, want bearer token", got)
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"health_status": "HEALTH_OK",
				"version":       "20.2.2",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer cephServer.Close()

	db := openDatabaseCephClientTestDB(t)
	if err := db.Create(&store.CephCluster{
		Name:                 "primary",
		Enabled:              true,
		DashboardEnabled:     true,
		DashboardBaseURL:     cephServer.URL,
		DashboardUsername:    "admin",
		DashboardPassword:    "password",
		DashboardInsecureTLS: true,
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
