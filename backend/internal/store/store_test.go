package store

import (
	"path/filepath"
	"testing"

	"cephtower/backend/internal/config"
)

func TestOpenSQLiteCreatesDatabaseAndMigrates(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "cephtower.db")
	db, err := Open(config.DatabaseConfig{
		Engine: EngineSQLite,
		SQLite: config.SQLiteConfig{
			Path: dbPath,
		},
	})
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	}()

	if !db.Migrator().HasTable(&Setting{}) {
		t.Fatal("settings table was not migrated")
	}
}
