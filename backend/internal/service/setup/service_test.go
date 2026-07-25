package setup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

func TestInitializeCreatesAdminAndReplacesDatabase(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("database:\n  engine: sqlite\n  sqlite:\n    path: initial.db\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	initial, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: "initial.db"}}, dir)
	if err != nil {
		t.Fatal(err)
	}
	manager := store.NewManager(initial)
	cfg := config.Config{Path: configPath, Server: config.ServerConfig{Dir: dir}, Database: config.DatabaseConfig{Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: "initial.db"}}}
	service := New(manager, func() config.Config { return cfg }, func(next config.Config) { cfg = next })
	err = service.Initialize(context.Background(), Input{Database: config.DatabaseConfig{Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: "new.db"}}, Username: "admin", Email: "admin@example.com", Password: "ChangeMe123!"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	user, err := manager.Current().FindUserByUsername(context.Background(), "admin")
	if err != nil || user.Role != store.UserRoleAdmin {
		t.Fatalf("admin = %#v, %v", user, err)
	}
}
