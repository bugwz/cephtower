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
	data := fmt.Sprintf("server:\n  address: 127.0.0.1\n  port: 36900\n  dir: %q\nlog:\n  output: stdout\ndatabase:\n  encryption_key: 0123456789abcdefghijklmnopqrstuv\n  engine: sqlite\n  sqlite:\n    path: data/cephtower.db\n", dir)
	if err := os.WriteFile(configPath, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	application, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := application.Close(ctx); err != nil {
		t.Fatal(err)
	}
}
