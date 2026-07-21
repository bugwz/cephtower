package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`http_addr: ":9090"
logging:
  level: debug
  format: json
database:
  engine: mysql
  mysql:
    host: db.example.com
    port: 3307
    username: cephtower
    password: db-secret
    database: cephtower
    params: charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9090")
	}
	if cfg.Path != path {
		t.Fatalf("Path = %q, want %q", cfg.Path, path)
	}
	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "json" {
		t.Fatalf("Logging = %#v, want debug/json", cfg.Logging)
	}
	if cfg.Database.Engine != "mysql" {
		t.Fatalf("Database.Engine = %q, want mysql", cfg.Database.Engine)
	}
	if cfg.Database.MySQL.Host != "db.example.com" || cfg.Database.MySQL.Port != 3307 {
		t.Fatalf("unexpected MySQL address: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Username != "cephtower" || cfg.Database.MySQL.Password != "db-secret" {
		t.Fatalf("unexpected MySQL credentials: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Database != "cephtower" {
		t.Fatalf("Database.MySQL.Database = %q, want cephtower", cfg.Database.MySQL.Database)
	}
	if cfg.Database.MySQL.Params != "charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai" {
		t.Fatalf("Database.MySQL.Params = %q, want configured params", cfg.Database.MySQL.Params)
	}
}

func TestSaveDatabaseRewritesDatabaseConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`http_addr: ":9090"
database:
  engine: sqlite
  sqlite:
    path: data/old.db
smtp:
  host: smtp.example.com
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	err := SaveDatabase(path, DatabaseConfig{
		Engine: "mysql",
		SQLite: SQLiteConfig{
			Path: "data/new.db",
		},
		MySQL: MySQLConfig{
			Host:     "db.example.com",
			Port:     3307,
			Username: "tower",
			Password: "secret",
			Database: "cephtower",
			Params:   "charset=utf8mb4",
		},
	})
	if err != nil {
		t.Fatalf("SaveDatabase() returned error: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.HTTPAddr != ":9090" || cfg.SMTP.Host != "smtp.example.com" {
		t.Fatalf("non-database fields were not preserved: %#v", cfg)
	}
	if cfg.Database.Engine != "mysql" {
		t.Fatalf("Database.Engine = %q, want mysql", cfg.Database.Engine)
	}
	if cfg.Database.SQLite.Path != "data/new.db" {
		t.Fatalf("Database.SQLite.Path = %q, want data/new.db", cfg.Database.SQLite.Path)
	}
	if cfg.Database.MySQL.Host != "db.example.com" || cfg.Database.MySQL.Port != 3307 {
		t.Fatalf("unexpected MySQL address: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Username != "tower" || cfg.Database.MySQL.Password != "secret" {
		t.Fatalf("unexpected MySQL credentials: %#v", cfg.Database.MySQL)
	}
}

func TestLoadDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`{}`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPAddr != ":36900" {
		t.Fatalf("HTTPAddr = %q, want default :36900", cfg.HTTPAddr)
	}
	if cfg.Database.Engine != "sqlite" {
		t.Fatalf("Database.Engine = %q, want default sqlite", cfg.Database.Engine)
	}
	if cfg.Database.SQLite.Path != "data/cephtower.db" {
		t.Fatalf("Database.SQLite.Path = %q, want default data/cephtower.db", cfg.Database.SQLite.Path)
	}
	if cfg.Database.MySQL.Host != "127.0.0.1" || cfg.Database.MySQL.Port != 3306 {
		t.Fatalf("unexpected default MySQL address: %#v", cfg.Database.MySQL)
	}
	if cfg.Database.MySQL.Params != "charset=utf8mb4&parseTime=True&loc=Local" {
		t.Fatalf("Database.MySQL.Params = %q, want default params", cfg.Database.MySQL.Params)
	}
	if cfg.Logging.Level != "info" || cfg.Logging.Format != "txt" {
		t.Fatalf("Logging defaults = %#v, want info/txt", cfg.Logging)
	}
}

func TestLoadRejectsUnsupportedDatabaseEngine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`database:
  engine: postgres
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error, want unsupported engine error")
	}
}

func TestLoadRejectsUnsupportedLoggingLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`logging:
  level: verbose
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error, want unsupported logging level error")
	}
}

func TestLoadRejectsUnsupportedLoggingFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`logging:
  format: xml
`)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("Load() returned nil error, want unsupported logging format error")
	}
}
