package store

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"cephtower/backend/internal/config"
	sqlitestore "cephtower/backend/internal/store/sqlite"
)

func TestTestInitializationTargetSQLiteAllowsMissingDatabase(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DatabaseConfig{Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "fresh.db"}}

	if err := TestInitializationTarget(context.Background(), cfg, workDir); err != nil {
		t.Fatalf("TestInitializationTarget() error = %v", err)
	}
	if _, err := os.Stat(sqlitestore.ResolveName("fresh.db", workDir)); !os.IsNotExist(err) {
		t.Fatalf("sqlite db was created during probe, stat err = %v", err)
	}
}

func TestTestInitializationTargetSQLiteRejectsExistingDatabase(t *testing.T) {
	workDir := t.TempDir()
	path := sqlitestore.ResolveName("existing.db", workDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create sqlite dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("not empty"), 0o600); err != nil {
		t.Fatalf("create sqlite db file: %v", err)
	}
	cfg := config.DatabaseConfig{Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "existing.db"}}

	err := TestInitializationTarget(context.Background(), cfg, workDir)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("TestInitializationTarget() error = %v, want already exists", err)
	}
}

func TestTestInitializationTargetSQLiteRejectsDatabaseDirectory(t *testing.T) {
	workDir := t.TempDir()
	path := sqlitestore.ResolveName("directory.db", workDir)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create sqlite db directory: %v", err)
	}
	cfg := config.DatabaseConfig{Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "directory.db"}}

	err := TestInitializationTarget(context.Background(), cfg, workDir)
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("TestInitializationTarget() error = %v, want directory error", err)
	}
}

func TestTestInitializationTargetSQLiteRejectsUnwritableDirectory(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission probe is not stable on this platform")
	}
	workDir := t.TempDir()
	dir := filepath.Join(workDir, "data", "db")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create sqlite dir: %v", err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("chmod sqlite dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})
	cfg := config.DatabaseConfig{Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "fresh.db"}}

	err := TestInitializationTarget(context.Background(), cfg, workDir)
	if err == nil || !strings.Contains(err.Error(), "not writable") {
		t.Fatalf("TestInitializationTarget() error = %v, want not writable", err)
	}
}
