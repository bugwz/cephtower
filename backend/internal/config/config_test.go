package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const validKey = "0123456789abcdefghijklmnopqrstuv"

func TestLoadAllowsEmptyDatabaseEncryptionKeyBeforeBootstrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("database:\n  encryption_key: \"\"\n  engine: sqlite\n  sqlite:\n    name: cephtower.db\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.EncryptionKey != "" {
		t.Fatal("empty pre-bootstrap key was not preserved")
	}
	if !cfg.Server.Bootstrap {
		t.Fatal("server.bootstrap should default to true")
	}
	if !cfg.Server.Auth {
		t.Fatal("server.auth should default to true")
	}
}

func TestLoadAcceptsDisabledServerAuth(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := "server:\n  auth: false\ndatabase:\n  encryption_key: \"\"\n  engine: sqlite\n  sqlite:\n    name: cephtower.db\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Auth {
		t.Fatal("server.auth was not disabled")
	}
}

func TestLoadAcceptsValidDatabaseEncryptionKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := "server:\n  dir: .\ndatabase:\n  encryption_key: " + validKey + "\n  engine: sqlite\n  sqlite:\n    name: cephtower.db\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.EncryptionKey != validKey {
		t.Fatal("key was not preserved")
	}
}

func TestSaveSetupWritesDatabaseAndDisablesBootstrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := "server:\n  dir: .\n  bootstrap: true\ndatabase:\n  encryption_key: \"\"\n  engine: sqlite\n  sqlite:\n    name: cephtower.db\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	database := DatabaseConfig{EncryptionKey: validKey, Engine: "mysql", SQLite: SQLiteConfig{Name: "ignored.db"}, MySQL: MySQLConfig{Host: "db.local", Port: 3307, Username: "tower", Password: "secret", Database: "tower", Params: "parseTime=True", TLS: "preferred"}}
	if err := SaveSetup(path, database, false); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Bootstrap {
		t.Fatal("server.bootstrap was not disabled")
	}
	if cfg.Database.Engine != "mysql" || cfg.Database.MySQL.Host != "db.local" || cfg.Database.MySQL.Database != "tower" {
		t.Fatalf("database was not saved: %+v", cfg.Database)
	}
}

func TestSaveSetupWritesSQLiteName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := "server:\n  dir: .\n  bootstrap: true\ndatabase:\n  encryption_key: \"\"\n  engine: sqlite\n  sqlite:\n    name: old.db\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	database := DatabaseConfig{EncryptionKey: validKey, Engine: "sqlite", SQLite: SQLiteConfig{Name: "cephtower.db"}}
	if err := SaveSetup(path, database, false); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output, []byte("name: cephtower.db")) {
		t.Fatalf("sqlite name was not written correctly:\n%s", output)
	}
}

func TestGenerateDatabaseEncryptionKeyProducesValidKey(t *testing.T) {
	key, err := GenerateDatabaseEncryptionKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("len(key) = %d", len(key))
	}
}
