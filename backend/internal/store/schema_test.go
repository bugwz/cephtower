package store

import (
	"cephtower/backend/internal/config"
	"context"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const schemaTestKey = "0123456789abcdefghijklmnopqrstuv"

func TestSQLiteBaselineContainsExactlyNineteenTables(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/schema.db"}})
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	var names []string
	if err := db.db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name").Scan(&names).Error; err != nil {
		t.Fatal(err)
	}
	expected := []string{"audit_event", "ceph_action_plan", "ceph_cluster", "ceph_cluster_capability", "ceph_cluster_credential", "ceph_cluster_endpoint", "ceph_cluster_observation", "ceph_collection_run", "ceph_operation", "ceph_operation_event", "ceph_operation_lock", "ceph_resource_record", "password_reset_code", "role", "schema_migration", "setting", "user", "user_role_binding", "user_session"}
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
	path := t.TempDir() + "/schema.db"
	cfg := config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Path: path}}
	first, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_ = Close(first)
	second, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_ = Close(second)
}

var expectedColumns = map[string][]string{
	"schema_migration":         {"version", "checksum", "applied_at"},
	"setting":                  {"key", "value", "created_at", "updated_at"},
	"user":                     {"id", "username", "display_name", "email", "password", "status", "last_login_at", "created_at", "updated_at"},
	"password_reset_code":      {"id", "user_id", "code_hash", "expires_at", "consumed_at", "created_at"},
	"user_session":             {"id", "user_id", "token_hash", "source_ip", "user_agent", "expires_at", "last_seen_at", "revoked_at", "created_at"},
	"role":                     {"id", "name", "description", "created_at", "updated_at"},
	"user_role_binding":        {"id", "user_id", "role_id", "cluster_id", "scope_key", "created_by_user_id", "created_at"},
	"ceph_cluster":             {"id", "name", "monitor_addresses", "client_username", "client_key", "created_at", "updated_at"},
	"ceph_cluster_observation": {"cluster_id", "fsid", "ceph_version", "status", "enabled", "generation", "last_seen_at", "last_error_code", "last_error_message", "observed_at", "updated_at"},
	"ceph_cluster_credential":  {"id", "cluster_id", "kind", "credential", "fingerprint", "created_at", "updated_at"},
	"ceph_cluster_endpoint":    {"id", "cluster_id", "kind", "name", "url", "tls_mode", "ca_credential_id", "config_json", "enabled", "created_at", "updated_at"},
	"ceph_cluster_capability":  {"id", "cluster_id", "name", "supported", "reason", "version", "details_json", "observed_at", "updated_at"},
	"ceph_resource_record":     {"id", "cluster_id", "kind", "natural_key", "parent_kind", "parent_key", "name", "status", "generation", "resource_version", "source", "source_version", "observed_at", "stale_at", "payload_schema_version", "payload_json", "created_at", "updated_at"},
	"ceph_collection_run":      {"id", "cluster_id", "module", "generation", "status", "source", "record_count", "error_code", "error_message", "started_at", "finished_at", "created_at"},
	"ceph_action_plan":         {"id", "cluster_id", "actor_user_id", "actor_username", "request_id", "action", "resource_kind", "resource_key", "resource_generation", "risk", "status", "request_json", "blockers_json", "warnings_json", "expires_at", "consumed_at", "created_at"},
	"ceph_operation":           {"id", "cluster_id", "cluster_name", "actor_user_id", "actor_username", "plan_id", "retry_of_id", "request_id", "action", "resource_kind", "resource_key", "resource_generation", "risk", "status", "stage", "progress", "attempt", "max_attempts", "idempotency_key_hash", "idempotency_scope_hash", "request_json", "result_json", "error_code", "error_message", "error_details_json", "retryable", "cancel_requested_at", "scheduled_at", "started_at", "heartbeat_at", "completed_at", "created_at", "updated_at"},
	"ceph_operation_event":     {"id", "operation_id", "sequence", "event_type", "stage", "progress", "message", "data_json", "error_code", "created_at"},
	"ceph_operation_lock":      {"lock_key", "cluster_id", "resource_kind", "resource_key", "operation_id", "fencing_token", "lease_expires_at", "acquired_at", "updated_at"},
	"audit_event":              {"id", "occurred_at", "event_type", "request_id", "actor_user_id", "actor_username", "cluster_id", "cluster_name", "action", "resource_kind", "resource_key", "risk", "outcome", "http_status", "error_code", "source_ip", "user_agent", "plan_id", "operation_id", "before_generation", "after_generation", "parameters_json", "details_json", "previous_hash", "event_hash"},
}

func TestSQLiteColumnsIndexesAndForeignKeysMatchBaseline(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/schema.db"}})
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
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
	for _, index := range []string{"idx_resource_parent", "idx_operation_heartbeat", "idx_audit_operation"} {
		var count int64
		if err := db.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name = ?", index).Scan(&count).Error; err != nil || count != 1 {
			t.Fatalf("index %s missing: count=%d err=%v", index, count, err)
		}
	}
}

func TestCanonicalDDLAndDocumentationContainExactlyBaselineTables(t *testing.T) {
	pattern := regexp.MustCompile(`(?im)^CREATE TABLE(?: IF NOT EXISTS)?\s+` + "`?" + `([a-z_]+)` + "`?" + `\s*\(`)
	expected := make([]string, 0, len(expectedColumns))
	for table := range expectedColumns {
		expected = append(expected, table)
	}
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
	for path, expectedDDL := range map[string]string{"../../../docs/database-schema/sqlite/schema.sql": sqliteBaselineSQL, "../../../docs/database-schema/mysql/schema.sql": mysqlBaselineSQL} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != expectedDDL {
			t.Fatalf("%s is not an exact canonical migration mirror", path)
		}
	}
	for _, forbidden := range []string{"password_ciphertext", "client_key_ciphertext", "credential_ciphertext", "nonce"} {
		if strings.Contains(sqliteBaselineSQL, forbidden) || strings.Contains(mysqlBaselineSQL, forbidden) {
			t.Fatalf("forbidden secret helper column %q exists", forbidden)
		}
	}
}
