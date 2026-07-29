package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewComposesOfflineRuntime(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	data := fmt.Sprintf("server:\n  address: 127.0.0.1\n  port: 36900\n  dir: %q\nlog:\n  output: stdout\ndatabase:\n  encryption_key: 0123456789abcdefghijklmnopqrstuv\n  engine: sqlite\n  sqlite:\n    name: cephtower.db\n", dir)
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	application, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data", "db", "cephtower.db")); !os.IsNotExist(err) {
		t.Fatalf("startup created sqlite database unexpectedly: %v", err)
	}
	for _, path := range []string{
		filepath.Join(dir, "log"),
		filepath.Join(dir, "data", "runtime"),
		filepath.Join(dir, "data", "db"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("startup did not create %s: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("startup path %s is not a directory", path)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := application.Close(ctx); err != nil {
		t.Fatal(err)
	}
}
