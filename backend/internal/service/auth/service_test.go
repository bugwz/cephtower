package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

const bootstrapKey = "0123456789abcdefghijklmnopqrstuv"

func TestInitialAdminCanOnlyBeCreatedOnce(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: bootstrapKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/auth.db"}})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	service := New(func() *store.Database { return db }, func() config.Config {
		return config.Config{Database: config.DatabaseConfig{EncryptionKey: bootstrapKey}}
	})
	if err := service.EnsureRoles(context.Background()); err != nil {
		t.Fatal(err)
	}
	required, err := service.BootstrapRequired(context.Background())
	if err != nil || !required {
		t.Fatalf("required=%v err=%v", required, err)
	}
	if _, err := service.CreateInitialAdmin(context.Background(), CreateUserInput{Username: "admin", DisplayName: "Admin", Password: "password-123"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateInitialAdmin(context.Background(), CreateUserInput{Username: "other", DisplayName: "Other", Password: "password-456"}); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("second bootstrap err=%v", err)
	}
	required, _ = service.BootstrapRequired(context.Background())
	if required {
		t.Fatal("bootstrap still required")
	}
}

func TestClusterRoleBindingLifecycle(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: bootstrapKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/rbac.db"}})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	service := New(func() *store.Database { return db }, func() config.Config {
		return config.Config{Database: config.DatabaseConfig{EncryptionKey: bootstrapKey}}
	})
	if err := service.EnsureRoles(context.Background()); err != nil {
		t.Fatal(err)
	}
	user, err := service.CreateUser(context.Background(), CreateUserInput{Username: "operator", DisplayName: "Operator", Password: "password-123", Role: "viewer"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.BindUserRole(context.Background(), user.ID, "storage-admin", &cluster.ID, nil); err != nil {
		t.Fatal(err)
	}
	rows, err := db.ListClusterRoleBindings(context.Background(), cluster.ID)
	if err != nil || len(rows) != 1 || rows[0].User.Username != "operator" || rows[0].Role.Name != "storage-admin" {
		t.Fatalf("bindings=%#v err=%v", rows, err)
	}
	if err := db.DeleteClusterRoleBinding(context.Background(), cluster.ID, rows[0].ID); err != nil {
		t.Fatal(err)
	}
	rows, err = db.ListClusterRoleBindings(context.Background(), cluster.ID)
	if err != nil || len(rows) != 0 {
		t.Fatalf("bindings after delete=%#v err=%v", rows, err)
	}
}
