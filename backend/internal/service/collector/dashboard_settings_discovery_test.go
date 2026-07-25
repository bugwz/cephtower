package collector

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

func TestSaveDiscoveredSettingsStoresRedactedSnapshots(t *testing.T) {
	db := openDiscoveryTestDB(t)
	defer closeDiscoveryTestDB(t, db)

	cluster := store.CephCluster{
		Name:              "primary",
		MonitorHost:       "10.0.0.1:6789",
		Keyring:           "keyring",
		DashboardUsername: "admin",
		DashboardPassword: "password",
	}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	count := saveDiscoveredSettings(context.Background(), db, cluster.ID, []byte(`[
		{"name":"GRAFANA_API_PASSWORD","type":"str","default":false,"value":"secret"},
		{"name":"PROMETHEUS_API_HOST","type":"str","default":false,"value":"http://prometheus:9090"}
	]`))
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}

	var password store.CephClusterSettingSnapshot
	if err := db.FindClusterRecord(context.Background(), cluster.ID, map[string]any{"name": "GRAFANA_API_PASSWORD"}, &password); err != nil {
		t.Fatalf("load password setting: %v", err)
	}
	if password.Group != "grafana" || !password.Sensitive || !password.ValueSet || password.ValueRedacted != `"********"` {
		t.Fatalf("password snapshot = %#v, want redacted grafana secret", password)
	}
	if containsPlainSecret(password.Payload, "secret") {
		t.Fatalf("payload leaked secret: %s", password.Payload)
	}

	var prometheus store.CephClusterSettingSnapshot
	if err := db.FindClusterRecord(context.Background(), cluster.ID, map[string]any{"name": "PROMETHEUS_API_HOST"}, &prometheus); err != nil {
		t.Fatalf("load prometheus setting: %v", err)
	}
	if prometheus.Group != "prometheus" || prometheus.Sensitive || prometheus.ValueRedacted != `"http://prometheus:9090"` {
		t.Fatalf("prometheus snapshot = %#v, want plain prometheus setting", prometheus)
	}
}

func TestSaveDiscoveredFeatureTogglesReplacesClusterToggles(t *testing.T) {
	db := openDiscoveryTestDB(t)
	defer closeDiscoveryTestDB(t, db)

	cluster := store.CephCluster{
		Name:              "primary",
		MonitorHost:       "10.0.0.1:6789",
		Keyring:           "keyring",
		DashboardUsername: "admin",
		DashboardPassword: "password",
	}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	count := saveDiscoveredFeatureToggles(context.Background(), db, cluster.ID, []byte(`{"rbd":true,"iscsi":false}`))
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	count = saveDiscoveredFeatureToggles(context.Background(), db, cluster.ID, []byte(`{"cephfs":true}`))
	if count != 1 {
		t.Fatalf("replacement count = %d, want 1", count)
	}

	var toggles []store.CephClusterFeatureToggle
	if err := db.ListClusterRecords(context.Background(), cluster.ID, "name asc", &toggles); err != nil {
		t.Fatalf("list toggles: %v", err)
	}
	if len(toggles) != 1 || toggles[0].Name != "cephfs" || !toggles[0].Enabled {
		t.Fatalf("toggles = %#v, want only enabled cephfs", toggles)
	}
}

func openDiscoveryTestDB(t *testing.T) *store.Database {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{
		Engine: store.EngineSQLite,
		SQLite: config.SQLiteConfig{Path: filepath.Join(t.TempDir(), "cephtower.db")},
	})
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	return db
}

func closeDiscoveryTestDB(t *testing.T, db *store.Database) {
	t.Helper()
	if err := store.Close(db); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func containsPlainSecret(payload string, secret string) bool {
	return strings.Contains(payload, secret)
}
