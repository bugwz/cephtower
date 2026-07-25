package auth

import (
	"context"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

func TestLoginCreatesUsableSession(t *testing.T) {
	db := openTestDatabase(t)
	hash, err := store.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatal(err)
	}
	user := store.User{Username: "admin", DisplayName: "Admin", Role: store.UserRoleAdmin, PasswordHash: hash, Enabled: true}
	if err := db.CreateUser(context.Background(), &user); err != nil {
		t.Fatal(err)
	}
	service := New(func() *store.Database { return db }, func() config.Config { return config.Config{} })
	result, err := service.Login(context.Background(), "admin", "ChangeMe123!")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	got, err := service.UserForToken(context.Background(), result.Token)
	if err != nil || got.ID != user.ID {
		t.Fatalf("UserForToken() = %#v, %v", got, err)
	}
}

func openTestDatabase(t *testing.T) *store.Database {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/auth.db"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	return db
}
