package setup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"cephtower/backend/internal/config"
	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/store"
)

func TestInitializedSetupRejectsBeforeInspectingTarget(t *testing.T) {
	service := &Service{Manager: store.NewManager(nil), CurrentConfig: func() config.Config { return config.Config{} }}
	if _, err := service.Initialize(context.Background(), config.DatabaseConfig{}, authservice.CreateUserInput{}); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("initialize: %v", err)
	}
	if err := service.TestDatabase(context.Background(), config.DatabaseConfig{}); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("database probe: %v", err)
	}
}

func TestOnlyOneConcurrentInitializationCanInstallDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("server:\n  bootstrap: true\ndatabase:\n  encryption_key: 0123456789abcdefghijklmnopqrstuv\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Path: path, Server: config.ServerConfig{Dir: dir, Bootstrap: true}, Database: config.DatabaseConfig{EncryptionKey: "0123456789abcdefghijklmnopqrstuv"}}
	manager := store.NewManager(nil)
	defer manager.Close()
	calls := 0
	service := &Service{Manager: manager, CurrentConfig: func() config.Config { return cfg }, UpdateConfig: func(next config.Config) { cfg = next }, OnInitialized: func() { calls++ }}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, name := range []string{"first.db", "second.db"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			<-start
			_, err := service.Initialize(context.Background(), config.DatabaseConfig{Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: name}}, authservice.CreateUserInput{Username: "admin", DisplayName: "Admin", Password: "password-123"})
			results <- err
		}(name)
	}
	close(start)
	wg.Wait()
	close(results)
	succeeded, rejected := 0, 0
	for err := range results {
		if err == nil {
			succeeded++
		} else if errors.Is(err, ErrAlreadyInitialized) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if succeeded != 1 || rejected != 1 || calls != 1 {
		t.Fatalf("success=%d rejected=%d callbacks=%d", succeeded, rejected, calls)
	}
	if err := service.TestDatabase(context.Background(), config.DatabaseConfig{}); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("probe after setup: %v", err)
	}
	if cfg.Server.Bootstrap {
		t.Fatal("bootstrap remains enabled")
	}
}
