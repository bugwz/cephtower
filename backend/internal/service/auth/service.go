package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const sessionTTL = 12 * time.Hour
const passwordResetTTL = 10 * time.Minute

var BuiltinRoles = []string{
	"viewer",
	"operator",
	"storage-admin",
	"cluster-admin",
	"security-admin",
}

var ErrAlreadyInitialized = errors.New("initial administrator already exists")

type Service struct {
	database      func() *store.Database
	currentConfig func() config.Config
}

func New(database func() *store.Database, currentConfig func() config.Config) *Service {
	return &Service{database: database, currentConfig: currentConfig}
}
func (s *Service) EnsureRoles(ctx context.Context) error {
	return s.database().EnsureBuiltinRoles(ctx, BuiltinRoles)
}

func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}
	user, err := s.database().FindUserByUsername(ctx, username)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	if user.Status != "active" {
		return LoginResult{}, ErrUserDisabled
	}
	plain, err := security.Decrypt(user.Password, s.currentConfig().Database.EncryptionKey)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	defer zero(plain)
	if !hmac.Equal(plain, []byte(password)) {
		return LoginResult{}, ErrInvalidCredentials
	}
	token, err := randomToken()
	if err != nil {
		return LoginResult{}, err
	}
	now := time.Now().UTC()
	session := store.UserSession{ID: uuid(), UserID: user.ID, TokenHash: store.SHA256(token), ExpiresAt: now.Add(sessionTTL), LastSeenAt: now, CreatedAt: now}
	if err := s.database().CreateSessionAndTouchUser(ctx, &session, now); err != nil {
		return LoginResult{}, err
	}
	user.LastLoginAt = &now
	return LoginResult{Token: token, ExpiresAt: session.ExpiresAt, User: user}, nil
}
func (s *Service) UserForToken(ctx context.Context, token string) (store.User, error) {
	if token == "" {
		return store.User{}, ErrInvalidCredentials
	}
	user, err := s.database().UserForSessionHash(ctx, store.SHA256(token), time.Now().UTC())
	if err != nil || user.Status != "active" {
		return store.User{}, ErrInvalidCredentials
	}
	return user, nil
}
func (s *Service) ListUsers(ctx context.Context) ([]store.User, error) {
	return s.database().ListUsers(ctx)
}
func (s *Service) BootstrapRequired(ctx context.Context) (bool, error) {
	count, err := s.database().CountUsers(ctx)
	return count == 0, err
}
func (s *Service) CreateInitialAdmin(ctx context.Context, input CreateUserInput) (store.User, error) {
	input.Role = "cluster-admin"
	var created store.User
	err := s.database().Transaction(func(tx *store.Database) error {
		now := time.Now().UTC()
		marker := store.Setting{Key: "auth.bootstrap", Value: "complete", CreatedAt: now, UpdatedAt: now}
		if err := tx.Insert(ctx, &marker); err != nil {
			return ErrAlreadyInitialized
		}
		count, err := tx.CountUsers(ctx)
		if err != nil {
			return err
		}
		if count != 0 {
			return ErrAlreadyInitialized
		}
		service := &Service{database: func() *store.Database { return tx }, currentConfig: s.currentConfig}
		created, err = service.createUser(ctx, input)
		return err
	})
	return created, err
}
func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (store.User, error) {
	return s.createUser(ctx, input)
}
func (s *Service) createUser(ctx context.Context, input CreateUserInput) (store.User, error) {
	username := strings.ToLower(strings.TrimSpace(input.Username))
	display := strings.TrimSpace(input.DisplayName)
	if username == "" || display == "" || len(input.Password) < 8 {
		return store.User{}, fmt.Errorf("username, display_name and password of at least 8 characters are required")
	}
	encrypted, err := security.Encrypt([]byte(input.Password), s.currentConfig().Database.EncryptionKey)
	if err != nil {
		return store.User{}, err
	}
	now := time.Now().UTC()
	status := "active"
	if input.Enabled != nil && !*input.Enabled {
		status = "disabled"
	}
	var email *string
	if value := strings.TrimSpace(input.Email); value != "" {
		email = &value
	}
	user := store.User{Username: username, DisplayName: display, Email: email, Password: encrypted, Status: status, CreatedAt: now, UpdatedAt: now}
	if err := s.database().Transaction(func(tx *store.Database) error {
		if err := tx.CreateUser(ctx, &user); err != nil {
			return err
		}
		role := input.Role
		if role == "" {
			role = "viewer"
		}
		return tx.BindUserRole(ctx, user.ID, role, nil, nil)
	}); err != nil {
		return store.User{}, err
	}
	return user, nil
}
func (s *Service) UpdateUser(ctx context.Context, id uint64, input UpdateUserInput) (store.User, error) {
	if _, err := s.database().FindUserByID(ctx, id); errors.Is(err, store.ErrRecordNotFound) {
		return store.User{}, ErrUserNotFound
	} else if err != nil {
		return store.User{}, err
	}
	updates := map[string]any{}
	if input.DisplayName != nil {
		value := strings.TrimSpace(*input.DisplayName)
		if value == "" {
			return store.User{}, fmt.Errorf("display_name is required")
		}
		updates["display_name"] = value
	}
	if input.Email != nil {
		value := strings.TrimSpace(*input.Email)
		if value == "" {
			updates["email"] = nil
		} else {
			updates["email"] = value
		}
	}
	if input.Password != nil {
		if len(*input.Password) < 8 {
			return store.User{}, fmt.Errorf("password must be at least 8 characters")
		}
		encrypted, err := security.Encrypt([]byte(*input.Password), s.currentConfig().Database.EncryptionKey)
		if err != nil {
			return store.User{}, err
		}
		updates["password"] = encrypted
	}
	if input.Enabled != nil {
		if *input.Enabled {
			updates["status"] = "active"
		} else {
			updates["status"] = "disabled"
		}
	}
	if len(updates) > 0 {
		if err := s.database().UpdateUser(ctx, id, updates); err != nil {
			return store.User{}, err
		}
	}
	if input.Enabled != nil && !*input.Enabled {
		_ = s.database().DeleteUserSessions(ctx, id)
	}
	return s.database().FindUserByID(ctx, id)
}
func (s *Service) RequestPasswordReset(ctx context.Context, account string) error {
	account = strings.TrimSpace(account)
	if account == "" {
		return fmt.Errorf("account is required")
	}
	user, err := s.database().FindUserByAccount(ctx, account)
	if errors.Is(err, store.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.Email == nil {
		return ErrEmailRequired
	}
	code, err := randomNumericCode(6)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	reset := store.PasswordResetCode{UserID: user.ID, CodeHash: store.SHA256(code), ExpiresAt: now.Add(passwordResetTTL), CreatedAt: now}
	return s.database().ReplacePasswordReset(ctx, &reset)
}
func (s *Service) ConfirmPasswordReset(ctx context.Context, account, code, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	user, err := s.database().FindUserByAccount(ctx, strings.TrimSpace(account))
	if err != nil {
		return ErrInvalidResetCode
	}
	reset, err := s.database().FindValidPasswordReset(ctx, user.ID, store.SHA256(strings.TrimSpace(code)), time.Now().UTC())
	if err != nil {
		return ErrInvalidResetCode
	}
	encrypted, err := security.Encrypt([]byte(newPassword), s.currentConfig().Database.EncryptionKey)
	if err != nil {
		return err
	}
	return s.database().CompletePasswordReset(ctx, user.ID, reset.ID, encrypted, time.Now().UTC())
}

func randomToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
func randomNumericCode(length int) (string, error) {
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}
func uuid() string {
	data := make([]byte, 16)
	_, _ = rand.Read(data)
	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", data[:4], data[4:6], data[6:8], data[8:10], data[10:])
}
func zero(value []byte) {
	for i := range value {
		value[i] = 0
	}
}

var _ = sha256.Size
