package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validKey = "0123456789abcdefghijklmnopqrstuv"

func TestLoadRequiresDatabaseEncryptionKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("database:\n  encryption_key: \"\"\n  engine: sqlite\n  sqlite:\n    path: ':memory:'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "encryption_key") {
		t.Fatalf("Load() error = %v", err)
	}
}
func TestLoadAcceptsValidDatabaseEncryptionKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := "server:\n  dir: .\ndatabase:\n  encryption_key: " + validKey + "\n  engine: sqlite\n  sqlite:\n    path: ':memory:'\n"
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
