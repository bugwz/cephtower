package store

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const baselineVersion = "20260802_dedicated_entity_tables_v1"
const monitorTablesVersion = "20260824_monitor_details_v1"
const cephAuthTablesVersion = "20260914_ceph_auth_v1"
const operationTableVersion = "20260922_ceph_operation_v1"

var cephAuthEntityKinds = []string{"ceph_user"}

var monitorEntityKinds = []string{"mon_perf_counter", "mon_status"}

const sqliteEntityTableDDL = `CREATE TABLE %s (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  natural_key TEXT NOT NULL,
  parent_kind TEXT NULL,
  parent_key TEXT NULL,
  name TEXT NULL,
  status TEXT NULL,
  generation INTEGER NOT NULL,
  resource_version INTEGER NOT NULL DEFAULT 1,
  source TEXT NOT NULL,
  source_version TEXT NULL,
  observed_at DATETIME NOT NULL,
  stale_at DATETIME NULL,
  configured_data TEXT NULL,
  discovered_data TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT (strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ','now')),
  updated_at DATETIME NOT NULL DEFAULT (strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ','now')),
  UNIQUE(cluster_id, natural_key)
)`

const mysqlEntityTableDDL = `CREATE TABLE %s (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  natural_key VARCHAR(512) NOT NULL,
  parent_kind VARCHAR(64) NULL,
  parent_key VARCHAR(512) NULL,
  name VARCHAR(512) NULL,
  status VARCHAR(64) NULL,
  generation BIGINT UNSIGNED NOT NULL,
  resource_version BIGINT UNSIGNED NOT NULL DEFAULT 1,
  source VARCHAR(32) NOT NULL,
  source_version VARCHAR(128) NULL,
  observed_at DATETIME(6) NOT NULL,
  stale_at DATETIME(6) NULL,
  configured_data LONGTEXT NULL,
  discovered_data LONGTEXT NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	UNIQUE KEY uq_ceph_%s_cluster_key(cluster_id,natural_key),
	INDEX idx_ceph_%s_name(cluster_id,name),
	INDEX idx_ceph_%s_status(cluster_id,status),
	INDEX idx_ceph_%s_generation(cluster_id,generation),
	INDEX idx_ceph_%s_observed(cluster_id,observed_at),
	INDEX idx_ceph_%s_parent(cluster_id,parent_kind,parent_key),
	INDEX idx_ceph_%s_stale(cluster_id,stale_at),
  FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE
) ENGINE=InnoDB`

const sqliteOperationTableDDL = `CREATE TABLE ceph_operation (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id INTEGER NOT NULL REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  actor_user_id INTEGER NULL REFERENCES user(id) ON DELETE SET NULL,
  request_id TEXT NOT NULL,
  idempotency_key TEXT NULL,
  action TEXT NOT NULL,
  resource_kind TEXT NOT NULL,
  resource_key TEXT NOT NULL,
  risk TEXT NOT NULL,
  lock_key TEXT NOT NULL,
  status TEXT NOT NULL,
  parameters_ciphertext TEXT NOT NULL,
  expected_version INTEGER NULL,
  result_json TEXT NULL,
  error_code TEXT NULL,
  error_message TEXT NULL,
  retryable INTEGER NOT NULL DEFAULT 0 CHECK(retryable IN (0,1)),
  attempts INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 1,
  next_attempt_at DATETIME NULL,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX uq_operation_idempotency ON ceph_operation(cluster_id, idempotency_key);
CREATE INDEX idx_operation_cluster_created ON ceph_operation(cluster_id, created_at);
CREATE INDEX idx_operation_actor_created ON ceph_operation(actor_user_id, created_at);
CREATE INDEX idx_operation_request ON ceph_operation(request_id);
CREATE INDEX idx_operation_status_next ON ceph_operation(status, next_attempt_at);`

const mysqlOperationTableDDL = `CREATE TABLE ceph_operation (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  cluster_id BIGINT UNSIGNED NOT NULL,
  actor_user_id BIGINT UNSIGNED NULL,
  request_id VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(128) NULL,
  action VARCHAR(128) NOT NULL,
  resource_kind VARCHAR(64) NOT NULL,
  resource_key VARCHAR(512) NOT NULL,
  risk VARCHAR(16) NOT NULL,
  lock_key VARCHAR(512) NOT NULL,
  status VARCHAR(32) NOT NULL,
  parameters_ciphertext LONGTEXT NOT NULL,
  expected_version BIGINT UNSIGNED NULL,
  result_json LONGTEXT NULL,
  error_code VARCHAR(64) NULL,
  error_message LONGTEXT NULL,
  retryable BOOLEAN NOT NULL DEFAULT FALSE,
  attempts INT UNSIGNED NOT NULL DEFAULT 0,
  max_attempts INT UNSIGNED NOT NULL DEFAULT 1,
  next_attempt_at DATETIME(6) NULL,
  started_at DATETIME(6) NULL,
  finished_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_operation_idempotency(cluster_id,idempotency_key),
  INDEX idx_operation_cluster_created(cluster_id,created_at),
  INDEX idx_operation_actor_created(actor_user_id,created_at),
  INDEX idx_operation_request(request_id),
  INDEX idx_operation_status_next(status,next_attempt_at),
  FOREIGN KEY(cluster_id) REFERENCES ceph_cluster(id) ON DELETE CASCADE,
  FOREIGN KEY(actor_user_id) REFERENCES user(id) ON DELETE SET NULL
) ENGINE=InnoDB`

//go:embed migrations/sqlite/init.sql
var sqliteBaselineSQL string

//go:embed migrations/mysql/init.sql
var mysqlBaselineSQL string

func migrate(db *gorm.DB) error {
	engine := db.Dialector.Name()
	var definition, registryDDL string
	switch engine {
	case EngineSQLite:
		definition = sqliteBaselineSQL
		registryDDL = "CREATE TABLE IF NOT EXISTS schema_migration (version TEXT PRIMARY KEY, checksum TEXT NOT NULL, applied_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')))"
	case EngineMySQL:
		definition = mysqlBaselineSQL
		registryDDL = "CREATE TABLE IF NOT EXISTS schema_migration (version VARCHAR(64) PRIMARY KEY, checksum VARCHAR(64) NOT NULL, applied_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)) ENGINE=InnoDB"
	default:
		return fmt.Errorf("unsupported migration engine %q", engine)
	}
	if err := db.Exec(registryDDL).Error; err != nil {
		return fmt.Errorf("create schema migration table: %w", err)
	}
	baselineKinds := entityKindsWithout(entityKinds, append(append([]string{}, monitorEntityKinds...), cephAuthEntityKinds...))
	entitySchema := strings.Join(baselineKinds, ",") + sqliteEntityTableDDL + mysqlEntityTableDDL
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(definition+entitySchema)))
	var applied SchemaMigration
	err := db.Where("version = ?", baselineVersion).First(&applied).Error
	if err == nil {
		if applied.Checksum != checksum {
			return fmt.Errorf("migration %s checksum mismatch", baselineVersion)
		}
		return migrateAdditionalEntityTables(db, engine)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("read migration registry: %w", err)
	}
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("inspect database before baseline: %w", err)
	}
	for _, table := range tables {
		if table != "schema_migration" && table != "sqlite_sequence" {
			return fmt.Errorf("unversioned table %q exists; rebuild the development database", table)
		}
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range strings.Split(definition, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("apply baseline statement: %w", err)
			}
		}
		if err := createEntityTables(tx, engine, baselineKinds); err != nil {
			return err
		}
		entry := SchemaMigration{Version: baselineVersion, Checksum: checksum, AppliedAt: time.Now().UTC()}
		if err := tx.Create(&entry).Error; err != nil {
			return fmt.Errorf("record migration: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return migrateAdditionalEntityTables(db, engine)
}

func migrateAdditionalEntityTables(db *gorm.DB, engine string) error {
	if err := migrateEntityTables(db, engine, monitorTablesVersion, monitorEntityKinds); err != nil {
		return err
	}
	if err := migrateEntityTables(db, engine, cephAuthTablesVersion, cephAuthEntityKinds); err != nil {
		return err
	}
	return migrateOperationTable(db, engine)
}

func migrateOperationTable(db *gorm.DB, engine string) error {
	definition := sqliteOperationTableDDL
	if engine == EngineMySQL {
		definition = mysqlOperationTableDDL
	} else if engine != EngineSQLite {
		return fmt.Errorf("unsupported migration engine %q", engine)
	}
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(definition)))
	var applied SchemaMigration
	err := db.Where("version = ?", operationTableVersion).First(&applied).Error
	if err == nil {
		if applied.Checksum != checksum {
			return fmt.Errorf("migration %s checksum mismatch", operationTableVersion)
		}
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("read migration registry: %w", err)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range strings.Split(definition, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("create operation table: %w", err)
			}
		}
		return tx.Create(&SchemaMigration{Version: operationTableVersion, Checksum: checksum, AppliedAt: time.Now().UTC()}).Error
	})
}

func migrateEntityTables(db *gorm.DB, engine, version string, kinds []string) error {
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(kinds, ",")+sqliteEntityTableDDL+mysqlEntityTableDDL)))
	var applied SchemaMigration
	err := db.Where("version = ?", version).First(&applied).Error
	if err == nil {
		if applied.Checksum != checksum {
			return fmt.Errorf("migration %s checksum mismatch", version)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("read migration registry: %w", err)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := createEntityTables(tx, engine, kinds); err != nil {
			return err
		}
		entry := SchemaMigration{Version: version, Checksum: checksum, AppliedAt: time.Now().UTC()}
		if err := tx.Create(&entry).Error; err != nil {
			return fmt.Errorf("record migration: %w", err)
		}
		return nil
	})
}

func createEntityTables(tx *gorm.DB, engine string, kinds []string) error {
	for _, kind := range kinds {
		table, _ := EntityTableName(kind)
		var statement string
		switch engine {
		case EngineSQLite:
			statement = fmt.Sprintf(sqliteEntityTableDDL, table)
		case EngineMySQL:
			statement = fmt.Sprintf(mysqlEntityTableDDL, table, kind, kind, kind, kind, kind, kind, kind)
		default:
			return fmt.Errorf("unsupported migration engine %q", engine)
		}
		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("create dedicated entity table %s: %w", table, err)
		}
		if engine == EngineSQLite {
			for suffix, columns := range map[string]string{
				"name":       "cluster_id,name",
				"status":     "cluster_id,status",
				"generation": "cluster_id,generation",
				"observed":   "cluster_id,observed_at",
				"parent":     "cluster_id,parent_kind,parent_key",
				"stale":      "cluster_id,stale_at",
			} {
				index := fmt.Sprintf("idx_ceph_%s_%s", kind, suffix)
				if err := tx.Exec(fmt.Sprintf("CREATE INDEX %s ON %s(%s)", index, table, columns)).Error; err != nil {
					return fmt.Errorf("create dedicated entity index %s: %w", index, err)
				}
			}
		}
	}
	return nil
}

func entityKindsWithout(values, excluded []string) []string {
	excludedSet := make(map[string]struct{}, len(excluded))
	for _, value := range excluded {
		excludedSet[value] = struct{}{}
	}
	result := make([]string, 0, len(values)-len(excluded))
	for _, value := range values {
		if _, skip := excludedSet[value]; !skip {
			result = append(result, value)
		}
	}
	return result
}
