package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: bootstrapKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "password.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	key := bootstrapKey
	service := New(func() *store.Database { return db }, func() config.Config { return config.Config{Database: config.DatabaseConfig{EncryptionKey: key}} })
	if err := service.EnsureRoles(ctx); err != nil {
		t.Fatal(err)
	}
	user, err := service.CreateInitialAdmin(ctx, CreateUserInput{Username: "admin", DisplayName: "Admin", Password: "initial-password"})
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("initial-password")); err != nil {
		t.Fatal(err)
	}
	if _, err := security.Decrypt(user.Password, key); err == nil {
		t.Fatal("password must not be decryptable")
	}
	key = "abcdefghijklmnopqrstuvwxyz012345"
	if _, err := service.Login(ctx, "admin", "initial-password"); err != nil {
		t.Fatalf("login depends on encryption key: %v", err)
	}
	if _, err := service.Login(ctx, "admin", "incorrect-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	updated := "updated-password"
	if _, err := service.UpdateUser(ctx, user.ID, UpdateUserInput{Password: &updated}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(ctx, "admin", "initial-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("old password still accepted")
	}
	if _, err := service.Login(ctx, "admin", updated); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	reset := store.PasswordResetCode{UserID: user.ID, CodeHash: store.SHA256("123456"), ExpiresAt: now.Add(time.Minute), CreatedAt: now}
	if err := db.ReplacePasswordReset(ctx, &reset); err != nil {
		t.Fatal(err)
	}
	if err := service.ConfirmPasswordReset(ctx, "admin", "123456", "reset-password"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(ctx, "admin", "reset-password"); err != nil {
		t.Fatal(err)
	}
	ciphertext, err := security.Encrypt([]byte("reset-password"), key)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateUser(ctx, user.ID, map[string]any{"password": ciphertext}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(ctx, "admin", "reset-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("encrypted password accepted")
	}
}

func TestPasswordHashBoundsAndSalt(t *testing.T) {
	for _, password := range []string{"", "1234567", strings.Repeat("a", 73), strings.Repeat("密", 25)} {
		if _, err := hashPassword(password); err == nil {
			t.Fatalf("accepted password with %d bytes", len(password))
		}
	}
	password := strings.Repeat("a", 72)
	first, err := hashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	second, err := hashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("hashes must use independent salts")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(first), []byte(password)); err != nil {
		t.Fatal(err)
	}
}
