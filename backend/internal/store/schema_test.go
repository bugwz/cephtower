package store

import (
	"cephtower/backend/internal/config"
	"context"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

const schemaTestKey = "0123456789abcdefghijklmnopqrstuv"

func TestSQLiteBaselineContainsExpectedTables(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "schema.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	var names []string
	if err := db.db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&names).Error; err != nil {
		t.Fatal(err)
	}
	expected := []string{"audit_event", "ceph_cluster", "ceph_cluster_capability", "ceph_cluster_credential", "ceph_cluster_endpoint", "ceph_collection_run", "ceph_host", "password_reset_code", "role", "schema_migration", "setting", "user", "user_role_binding", "user_session"}
	for _, kind := range EntityKinds() {
		table, _ := EntityTableName(kind)
		expected = append(expected, table)
	}
	sort.Strings(expected)
	if len(names) != len(expected) {
		t.Fatalf("tables = %v", names)
	}
	for i := range expected {
		if names[i] != expected[i] {
			t.Fatalf("tables = %v", names)
		}
	}
	if err := db.Insert(context.Background(), &CephCluster{Name: "c", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher"}); err != nil {
		t.Fatal(err)
	}
}
func TestMigrationRegistryIsIdempotent(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "schema.db"}}
	first, err := Open(cfg, workDir)
	if err != nil {
		t.Fatal(err)
	}
	_ = Close(first)
	second, err := Open(cfg, workDir)
	if err != nil {
		t.Fatal(err)
	}
	_ = Close(second)
}

func TestClusterDiscoveryFSIDIsNotUnique(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "schema.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	now := time.Now().UTC()
	first := CephCluster{Name: "first", MonitorAddresses: "mon:6789", ClientUsername: "client.first", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	second := CephCluster{Name: "second", MonitorAddresses: "mon:6789", ClientUsername: "client.second", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &first); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateCluster(context.Background(), &second); err != nil {
		t.Fatal(err)
	}
	fsid := "9f11d824-8e53-11f1-8c55-00163e11e24f"
	for _, cluster := range []CephCluster{first, second} {
		cluster.FSID, cluster.Status, cluster.Enabled, cluster.DiscoveredData = &fsid, "available", true, `{}`
		if err := db.UpdateClusterDiscovery(context.Background(), cluster); err != nil {
			t.Fatalf("update discovery for cluster %d: %v", cluster.ID, err)
		}
	}
}

var expectedColumns = map[string][]string{
	"schema_migration":        {"version", "checksum", "applied_at"},
	"setting":                 {"key", "value", "created_at", "updated_at"},
	"user":                    {"id", "username", "display_name", "email", "password", "status", "last_login_at", "created_at", "updated_at"},
	"password_reset_code":     {"id", "user_id", "code_hash", "expires_at", "consumed_at", "created_at"},
	"user_session":            {"id", "user_id", "token_hash", "source_ip", "user_agent", "expires_at", "last_seen_at", "revoked_at", "created_at"},
	"role":                    {"id", "name", "description", "created_at", "updated_at"},
	"user_role_binding":       {"id", "user_id", "role_id", "cluster_id", "scope_key", "created_by_user_id", "created_at"},
	"ceph_cluster":            {"id", "name", "monitor_addresses", "client_username", "client_key", "discovered_data", "fsid", "ceph_version", "status", "enabled", "generation", "last_seen_at", "last_error_code", "last_error_message", "observed_at", "created_at", "updated_at"},
	"ceph_cluster_credential": {"id", "cluster_id", "kind", "credential", "fingerprint", "created_at", "updated_at"},
	"ceph_cluster_endpoint":   {"id", "cluster_id", "kind", "name", "url", "tls_mode", "ca_credential_id", "config_json", "enabled", "created_at", "updated_at"},
	"ceph_cluster_capability": {"id", "cluster_id", "name", "supported", "reason", "version", "details_json", "observed_at", "updated_at"},
	"ceph_host":               {"id", "cluster_id", "hostname", "ssh_address", "ssh_port", "ssh_user", "ssh_password_secret", "address", "status", "configured_data", "discovered_data", "generation", "resource_version", "source", "source_version", "observed_at", "stale_at", "created_at", "updated_at"},
	"ceph_collection_run":     {"id", "cluster_id", "module", "generation", "status", "source", "record_count", "error_code", "error_message", "started_at", "finished_at", "created_at"},
	"audit_event":             {"id", "occurred_at", "event_type", "request_id", "actor_user_id", "actor_username", "cluster_id", "cluster_name", "action", "resource_kind", "resource_key", "risk", "outcome", "http_status", "error_code", "source_ip", "user_agent", "before_generation", "after_generation", "parameters_json", "details_json", "previous_hash", "event_hash"},
}

var expectedEntityColumns = []string{"id", "cluster_id", "natural_key", "parent_kind", "parent_key", "name", "status", "generation", "resource_version", "source", "source_version", "observed_at", "stale_at", "configured_data", "discovered_data", "created_at", "updated_at"}

func TestSQLiteColumnsIndexesAndForeignKeysMatchBaseline(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "schema.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	for _, kind := range EntityKinds() {
		table, _ := EntityTableName(kind)
		expectedColumns[table] = expectedEntityColumns
	}
	for table, expected := range expectedColumns {
		var rows []struct {
			Name string `gorm:"column:name"`
		}
		if err := db.db.Raw("PRAGMA table_info('" + table + "')").Scan(&rows).Error; err != nil {
			t.Fatal(err)
		}
		actual := make([]string, 0, len(rows))
		for _, row := range rows {
			actual = append(actual, row.Name)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("%s columns = %v, want %v", table, actual, expected)
		}
	}
	var foreignKeys int64
	if err := db.db.Raw("SELECT COUNT(*) FROM pragma_foreign_key_list('ceph_cluster_endpoint') WHERE `table` = 'ceph_cluster_credential' AND on_delete = 'SET NULL'").Scan(&foreignKeys).Error; err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatal("endpoint CA foreign key missing")
	}
	for _, index := range []string{"idx_ceph_osd_parent", "idx_audit_request"} {
		var count int64
		if err := db.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name = ?", index).Scan(&count).Error; err != nil || count != 1 {
			t.Fatalf("index %s missing: count=%d err=%v", index, count, err)
		}
	}
}

func TestCanonicalDDLContainsExactlyBaselineTables(t *testing.T) {
	pattern := regexp.MustCompile(`(?im)^CREATE TABLE(?: IF NOT EXISTS)?\s+` + "`?" + `([a-z_]+)` + "`?" + `\s*\(`)
	expected := []string{"audit_event", "ceph_cluster", "ceph_cluster_capability", "ceph_cluster_credential", "ceph_cluster_endpoint", "ceph_collection_run", "ceph_host", "password_reset_code", "role", "schema_migration", "setting", "user", "user_role_binding", "user_session"}
	sort.Strings(expected)
	for name, ddl := range map[string]string{"sqlite": sqliteBaselineSQL, "mysql": mysqlBaselineSQL} {
		matches := pattern.FindAllStringSubmatch(ddl, -1)
		actual := make([]string, 0, len(matches))
		for _, match := range matches {
			actual = append(actual, match[1])
		}
		sort.Strings(actual)
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("%s DDL tables = %v", name, actual)
		}
		if strings.Count(ddl, "CREATE TABLE") != strings.Count(ddl, ";")-strings.Count(ddl, "CREATE INDEX") {
			t.Fatalf("%s DDL has an unexpected statement boundary", name)
		}
	}
	for _, forbidden := range []string{"password_ciphertext", "client_key_ciphertext", "credential_ciphertext", "nonce"} {
		if strings.Contains(sqliteBaselineSQL, forbidden) || strings.Contains(mysqlBaselineSQL, forbidden) {
			t.Fatalf("forbidden secret helper column %q exists", forbidden)
		}
	}
}
