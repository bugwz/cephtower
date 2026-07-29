package store

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const baselineVersion = "20260726_ceph_backend_v1"

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
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(definition)))
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
		entry := SchemaMigration{Version: baselineVersion, Checksum: checksum, AppliedAt: time.Now().UTC()}
		if err := tx.Create(&entry).Error; err != nil {
			return fmt.Errorf("record migration: %w", err)
		}
		return nil
	})
}
