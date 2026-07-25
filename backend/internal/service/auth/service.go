package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/smtp"
	"strings"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

const sessionTTL = 12 * time.Hour
const passwordResetTTL = 10 * time.Minute

type Service struct {
	database      func() *store.Database
	currentConfig func() config.Config
}

func New(database func() *store.Database, currentConfig func() config.Config) *Service {
	return &Service{database: database, currentConfig: currentConfig}
}

func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return LoginResult{}, fmt.Errorf("username and password are required")
	}
	user, err := s.database().FindUserByUsername(ctx, username)
	if errors.Is(err, store.ErrRecordNotFound) || (err == nil && !store.CheckPassword(password, user.PasswordHash)) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}
	if !user.Enabled {
		return LoginResult{}, ErrUserDisabled
	}
	token, err := randomToken()
	if err != nil {
		return LoginResult{}, err
	}
	now := time.Now().UTC()
	session := store.UserSession{Token: token, UserID: user.ID, ExpiresAt: now.Add(sessionTTL)}
	user.LastLoginAt = &now
	if err := s.database().CreateSessionAndTouchUser(ctx, &session, &user, now); err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token, ExpiresAt: session.ExpiresAt, User: user}, nil
}

func (s *Service) UserForToken(ctx context.Context, token string) (store.User, error) {
	if strings.TrimSpace(token) == "" {
		return store.User{}, ErrInvalidCredentials
	}
	user, err := s.database().UserForSession(ctx, token, time.Now().UTC())
	if err != nil || !user.Enabled {
		return store.User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]store.User, error) {
	return s.database().ListUsers(ctx)
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (store.User, error) {
	user, err := buildUser(input)
	if err != nil {
		return store.User{}, err
	}
	if err := s.database().CreateUser(ctx, &user); err != nil {
		return store.User{}, err
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uint, input UpdateUserInput) (store.User, error) {
	user, err := s.database().FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return store.User{}, ErrUserNotFound
		}
		return store.User{}, err
	}
	updates := map[string]any{}
	if input.DisplayName != nil {
		displayName := strings.TrimSpace(*input.DisplayName)
		if displayName == "" {
			return store.User{}, fmt.Errorf("display_name is required")
		}
		updates["display_name"] = displayName
	}
	if input.Email != nil {
		updates["email"] = strings.TrimSpace(*input.Email)
	}
	if input.Role != nil {
		role, err := normalizeRole(*input.Role)
		if err != nil {
			return store.User{}, err
		}
		updates["role"] = role
		if input.Permissions == nil {
			updates["permissions"] = permissionsJSON(nil, role)
		}
	}
	if input.Permissions != nil {
		updates["permissions"] = permissionsJSON(input.Permissions, user.Role)
	}
	if input.Password != nil {
		if len(*input.Password) < 8 {
			return store.User{}, fmt.Errorf("password must be at least 8 characters")
		}
		passwordHash, err := store.HashPassword(*input.Password)
		if err != nil {
			return store.User{}, err
		}
		updates["password_hash"] = passwordHash
	}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if len(updates) > 0 {
		if err := s.database().UpdateUser(ctx, &user, updates); err != nil {
			return store.User{}, err
		}
		if input.Enabled != nil && !*input.Enabled {
			if err := s.database().DeleteUserSessions(ctx, user.ID); err != nil {
				return store.User{}, err
			}
		}
		if user, err = s.database().FindUserByID(ctx, id); err != nil {
			return store.User{}, err
		}
	}
	return user, nil
}

// RequestPasswordReset intentionally succeeds when the account does not exist.
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
	if strings.TrimSpace(user.Email) == "" {
		return ErrEmailRequired
	}
	code, err := randomNumericCode(6)
	if err != nil {
		return err
	}
	codeHash, err := store.HashPassword(code)
	if err != nil {
		return err
	}
	reset := store.PasswordResetCode{UserID: user.ID, CodeHash: codeHash, ExpiresAt: time.Now().UTC().Add(passwordResetTTL)}
	if err := s.database().ReplacePasswordReset(ctx, &reset); err != nil {
		return err
	}
	return s.sendPasswordResetCode(user, code)
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, account, code, newPassword string) error {
	account = strings.TrimSpace(account)
	code = strings.TrimSpace(code)
	if account == "" || code == "" || newPassword == "" {
		return fmt.Errorf("account, code and new_password are required")
	}
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	user, err := s.database().FindUserByAccount(ctx, account)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return ErrInvalidResetCode
		}
		return err
	}
	reset, err := s.database().FindValidPasswordReset(ctx, user.ID, time.Now().UTC())
	if err != nil || !store.CheckPassword(code, reset.CodeHash) {
		return ErrInvalidResetCode
	}
	passwordHash, err := store.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.database().CompletePasswordReset(ctx, &user, &reset, passwordHash)
}

func buildUser(input CreateUserInput) (store.User, error) {
	username := strings.TrimSpace(input.Username)
	displayName := strings.TrimSpace(input.DisplayName)
	if username == "" || displayName == "" || input.Password == "" {
		return store.User{}, fmt.Errorf("username, display_name and password are required")
	}
	if len(input.Password) < 8 {
		return store.User{}, fmt.Errorf("password must be at least 8 characters")
	}
	role, err := normalizeRole(input.Role)
	if err != nil {
		return store.User{}, err
	}
	passwordHash, err := store.HashPassword(input.Password)
	if err != nil {
		return store.User{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	return store.User{
		Username: username, DisplayName: displayName, Email: strings.TrimSpace(input.Email), Role: role,
		Permissions: permissionsJSON(input.Permissions, role), PasswordHash: passwordHash, Enabled: enabled,
	}, nil
}

func normalizeRole(role string) (string, error) {
	switch strings.TrimSpace(role) {
	case "", store.UserRoleUser:
		return store.UserRoleUser, nil
	case store.UserRoleAdmin:
		return store.UserRoleAdmin, nil
	default:
		return "", fmt.Errorf("role must be admin or user")
	}
}

func permissionsJSON(permissions []string, role string) string {
	if permissions == nil {
		if role == store.UserRoleAdmin {
			permissions = []string{"cluster:read", "storage:read", "system:read", "user:manage"}
		} else {
			permissions = []string{"cluster:read", "storage:read"}
		}
	}
	payload, err := json.Marshal(permissions)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func PermissionsJSON(permissions []string, role string) string {
	return permissionsJSON(permissions, role)
}

func (s *Service) sendPasswordResetCode(user store.User, code string) error {
	cfg := s.currentConfig()
	if strings.TrimSpace(cfg.SMTP.Host) == "" {
		slog.Info("cephtower password reset code", "username", user.Username, "email", user.Email, "code", code)
		return nil
	}
	port := cfg.SMTP.Port
	if port == 0 {
		port = 587
	}
	from := cfg.SMTP.From
	if from == "" {
		from = cfg.SMTP.Username
	}
	if from == "" {
		return fmt.Errorf("smtp from address is required")
	}
	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, port)
	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	body := fmt.Sprintf("您的 CephTower 密码重置验证码是：%s\n\n验证码将在 %d 分钟后过期。", code, int(passwordResetTTL.Minutes()))
	message := strings.Join([]string{
		"From: " + from, "To: " + user.Email, "Subject: CephTower 密码重置验证码",
		"MIME-Version: 1.0", "Content-Type: text/plain; charset=UTF-8", "", body,
	}, "\r\n")
	if err := smtp.SendMail(addr, auth, from, []string{user.Email}, []byte(message)); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}

func randomToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func randomNumericCode(length int) (string, error) {
	var builder strings.Builder
	for i := 0; i < length; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate reset code: %w", err)
		}
		builder.WriteString(value.String())
	}
	return builder.String(), nil
}
