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
	if !db.Migrator().HasTable(&User{}) {
		t.Fatal("users table was not migrated")
	}
	if !db.Migrator().HasTable(&UserSession{}) {
		t.Fatal("user_sessions table was not migrated")
	}
	if !db.Migrator().HasTable(&PasswordResetCode{}) {
		t.Fatal("password_reset_codes table was not migrated")
	}

	var users int64
	if err := db.Model(&User{}).Count(&users).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 0 {
		t.Fatalf("users = %d, want empty database before setup", users)
	}
}
