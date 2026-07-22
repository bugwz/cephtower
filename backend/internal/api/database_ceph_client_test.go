package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integrations/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func TestDatabaseCephClientUsesLocalClusterSummary(t *testing.T) {
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
	if err := db.Create(&store.CephClusterSummary{
		ClusterID:         cluster.ID,
		HealthStatus:      "HEALTH_OK",
		Version:           "20.2.2",
		MgrID:             "mgr-a",
		MgrHost:           "node-a",
		HaveMonConnection: true,
		ExecutingTasks:    `[]`,
		FinishedTasks:     `[]`,
		Payload:           `{"health_status":"HEALTH_OK","version":"20.2.2"}`,
		DiscoveredAt:      time.Now(),
	}).Error; err != nil {
		t.Fatalf("create cluster summary: %v", err)
	}

	client := newDatabaseCephClient(func() *gorm.DB { return db })
	summary, err := client.ClusterSummary(context.Background())
	if err != nil {
		t.Fatalf("ClusterSummary() returned error: %v", err)
	}
	if summary.HealthStatus != "HEALTH_OK" || summary.Version != "20.2.2" {
		t.Fatalf("summary = %#v, want local database response", summary)
	}
}

func TestDatabaseCephClientListsHostsFromDatabase(t *testing.T) {
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
	if err := db.Create(&store.CephClusterHost{
		ClusterID:    cluster.ID,
		Hostname:     "node-a",
		Addr:         "10.0.0.1",
		Labels:       `[]`,
		Sources:      `{"ceph":true,"orchestrator":true}`,
		Payload:      `{"hostname":"node-a","addr":"10.0.0.1"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create host: %v", err)
	}
	if err := db.Create(&store.CephClusterHost{
		ClusterID:    cluster.ID,
		Hostname:     "node-b",
		Addr:         "10.0.0.2",
		Labels:       `[]`,
		Sources:      `{"ceph":true,"orchestrator":true}`,
		Payload:      `{"hostname":"node-b","addr":"10.0.0.2"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create host: %v", err)
	}

	client := newDatabaseCephClient(func() *gorm.DB { return db })
	hosts, err := client.ListHosts(context.Background(), ceph.ListHostsOptions{Search: "node-b"})
	if err != nil {
		t.Fatalf("ListHosts() returned error: %v", err)
	}
	if len(hosts) != 1 || hosts[0].Hostname != "node-b" {
		t.Fatalf("hosts = %#v, want node-b from database", hosts)
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

func TestDatabaseCephClientReturnsUnknownHealthWithoutDashboardCluster(t *testing.T) {
	db := openDatabaseCephClientTestDB(t)
	client := newDatabaseCephClient(func() *gorm.DB { return db })

	health, err := client.HealthFull(context.Background())
	if err != nil {
		t.Fatalf("HealthFull() returned error: %v", err)
	}
	healthBlock, ok := health["health"].(map[string]any)
	if !ok {
		t.Fatalf("health = %#v, want health block", health)
	}
	if healthBlock["status"] != "unknown" {
		t.Fatalf("status = %#v, want unknown", healthBlock["status"])
	}
}

func TestDatabaseCephClientRawMonitorReturnsEmptyWithoutCluster(t *testing.T) {
	db := openDatabaseCephClientTestDB(t)
	client := newDatabaseCephClient(func() *gorm.DB { return db })

	payload, err := client.Raw(context.Background(), http.MethodGet, "/api/monitor", nil, nil)
	if err != nil {
		t.Fatalf("Raw() returned error: %v", err)
	}

	var response map[string][]map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal monitor payload: %v", err)
	}
	if len(response["in_quorum"]) != 0 || len(response["out_quorum"]) != 0 || len(response["mons"]) != 0 {
		t.Fatalf("response = %#v, want empty monitor lists", response)
	}
}

func TestDatabaseCephClientRawLocalGETsReturnEmptyWithoutCluster(t *testing.T) {
	db := openDatabaseCephClientTestDB(t)
	client := newDatabaseCephClient(func() *gorm.DB { return db })
	paths := []string{
		"/api/service",
		"/api/service/known_types",
		"/api/service/mds/daemons",
		"/api/mgr/module",
		"/api/pool",
		"/api/pool/rbd/configuration",
		"/api/block/image",
		"/api/block/image/default_features",
		"/api/block/image/clone_format_version",
		"/api/block/image/trash",
		"/api/block/mirroring/summary",
		"/api/cephfs",
		"/api/cephfs/1",
		"/api/rgw/daemon",
		"/api/rgw/daemon/rgw-a",
		"/api/rgw/user",
		"/api/rgw/user/demo",
		"/api/rgw/bucket",
		"/api/rgw/bucket/demo",
		"/api/rgw/accounts",
		"/api/rgw/accounts/demo",
		"/api/cluster_conf",
		"/api/cluster_conf/filter",
		"/api/cluster_conf/osd_pool_default_size",
		"/api/logs/all",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			payload, err := client.Raw(context.Background(), http.MethodGet, path, nil, nil)
			if err != nil {
				t.Fatalf("Raw() returned error: %v", err)
			}
			if !json.Valid(payload) {
				t.Fatalf("Raw() returned invalid JSON: %s", payload)
			}
		})
	}
}

func TestDatabaseCephClientRawMonitorUsesDatabaseRecords(t *testing.T) {
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
	if err := db.Create(&store.CephClusterMon{
		ClusterID:    cluster.ID,
		Name:         "mon-a",
		Rank:         "0",
		Addr:         "10.0.0.11:6789",
		PublicAddr:   "10.0.0.11:6789",
		Status:       "in_quorum",
		Payload:      `{"name":"mon-a","rank":0,"addr":"10.0.0.11:6789"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create in-quorum mon: %v", err)
	}
	if err := db.Create(&store.CephClusterMon{
		ClusterID:    cluster.ID,
		Name:         "mon-b",
		Rank:         "1",
		Addr:         "10.0.0.12:6789",
		PublicAddr:   "10.0.0.12:6789",
		Status:       "out_quorum",
		Payload:      `{"name":"mon-b","rank":1,"addr":"10.0.0.12:6789"}`,
		DiscoveredAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("create out-quorum mon: %v", err)
	}

	client := newDatabaseCephClient(func() *gorm.DB { return db })
	payload, err := client.Raw(context.Background(), http.MethodGet, "/api/monitor", nil, nil)
	if err != nil {
		t.Fatalf("Raw() returned error: %v", err)
	}

	var response map[string][]map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal monitor payload: %v", err)
	}
	if len(response["in_quorum"]) != 1 || response["in_quorum"][0]["name"] != "mon-a" {
		t.Fatalf("in_quorum = %#v, want mon-a", response["in_quorum"])
	}
	if len(response["out_quorum"]) != 1 || response["out_quorum"][0]["name"] != "mon-b" {
		t.Fatalf("out_quorum = %#v, want mon-b", response["out_quorum"])
	}
	if len(response["mons"]) != 2 {
		t.Fatalf("mons = %#v, want both monitors", response["mons"])
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
