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
	entitySchema := strings.Join(entityKinds, ",") + sqliteEntityTableDDL + mysqlEntityTableDDL
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(definition+entitySchema)))
	var applied SchemaMigration
	err := db.Where("version = ?", baselineVersion).First(&applied).Error
	if err == nil {
		if applied.Checksum != checksum {
			return fmt.Errorf("migration %s checksum mismatch", baselineVersion)
		}
		return nil
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
	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range strings.Split(definition, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("apply baseline statement: %w", err)
			}
		}
		if err := createEntityTables(tx, engine); err != nil {
			return err
		}
		entry := SchemaMigration{Version: baselineVersion, Checksum: checksum, AppliedAt: time.Now().UTC()}
		if err := tx.Create(&entry).Error; err != nil {
			return fmt.Errorf("record migration: %w", err)
		}
		return nil
	})
}

func createEntityTables(tx *gorm.DB, engine string) error {
	for _, kind := range entityKinds {
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
