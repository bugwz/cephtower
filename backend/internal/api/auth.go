package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

const sessionTTL = 12 * time.Hour
const passwordResetTTL = 10 * time.Minute

type userResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
	Enabled     bool       `json:"enabled"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/password-reset/request", s.requestPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/password-reset/confirm", s.confirmPasswordReset)
	mux.HandleFunc("GET /api/v1/auth/me", s.me)
	mux.HandleFunc("GET /api/v1/users", s.listUsers)
	mux.HandleFunc("POST /api/v1/users", s.createUser)
	mux.HandleFunc("PATCH /api/v1/users/{id}", s.updateUser)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password are required"})
		return
	}

	var user store.User
	db := s.database()
	err := db.Where("username = ?", req.Username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !store.CheckPassword(req.Password, user.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
		return
	}
	if !user.Enabled {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "user is disabled"})
		return
	}

	token, err := randomToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	session := store.UserSession{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: now.Add(sessionTTL),
	}
	user.LastLoginAt = &now
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&session).Error; err != nil {
			return err
		}
		return tx.Model(&user).Update("last_login_at", now).Error
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": session.ExpiresAt,
		"user":       toUserResponse(user),
	})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var users []store.User
	if err := s.database().Order("id asc").Find(&users).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := make([]userResponse, 0, len(users))
	for _, user := range users {
		response = append(response, toUserResponse(user))
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var req struct {
		Username    string   `json:"username"`
		DisplayName string   `json:"display_name"`
		Email       string   `json:"email"`
		Role        string   `json:"role"`
		Permissions []string `json:"permissions"`
		Password    string   `json:"password"`
		Enabled     *bool    `json:"enabled"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := buildUser(req.Username, req.DisplayName, req.Email, req.Role, req.Permissions, req.Password, req.Enabled)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.database().Create(&user).Error; err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req struct {
		DisplayName *string  `json:"display_name"`
		Email       *string  `json:"email"`
		Role        *string  `json:"role"`
		Permissions []string `json:"permissions"`
		Password    *string  `json:"password"`
		Enabled     *bool    `json:"enabled"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	var user store.User
	db := s.database()
	if err := db.First(&user, id).Error; err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	updates := map[string]any{}
	if req.DisplayName != nil {
		displayName := strings.TrimSpace(*req.DisplayName)
		if displayName == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "display_name is required"})
			return
		}
		updates["display_name"] = displayName
	}
	if req.Email != nil {
		updates["email"] = strings.TrimSpace(*req.Email)
	}
	if req.Role != nil {
		role, err := normalizeRole(*req.Role)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		updates["role"] = role
		if req.Permissions == nil {
			updates["permissions"] = permissionsJSON(nil, role)
		}
	}
	if req.Permissions != nil {
		updates["permissions"] = permissionsJSON(req.Permissions, user.Role)
	}
	if req.Password != nil {
		if len(*req.Password) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
			return
		}
		passwordHash, err := store.HashPassword(*req.Password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		updates["password_hash"] = passwordHash
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if len(updates) == 0 {
		writeJSON(w, http.StatusOK, toUserResponse(user))
		return
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if req.Enabled != nil && !*req.Enabled {
		_ = db.Where("user_id = ?", user.ID).Delete(&store.UserSession{}).Error
	}
	if err := db.First(&user, id).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *Server) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Account string `json:"account"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	account := strings.TrimSpace(req.Account)
	if account == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account is required"})
		return
	}

	var user store.User
	db := s.database()
	err := db.Where("username = ? OR email = ?", account, account).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "如果账号存在，验证码将发送到绑定邮箱"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if strings.TrimSpace(user.Email) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前账号未绑定邮箱，请联系管理员重设密码"})
		return
	}

	code, err := randomNumericCode(6)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	codeHash, err := store.HashPassword(code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	reset := store.PasswordResetCode{
		UserID:    user.ID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().UTC().Add(passwordResetTTL),
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&store.PasswordResetCode{}).Where("user_id = ? AND used = ?", user.ID, false).Update("used", true).Error; err != nil {
			return err
		}
		return tx.Create(&reset).Error
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := s.sendPasswordResetCode(user, code); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "验证码已发送，请查收邮箱"})
}

func (s *Server) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Account     string `json:"account"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	account := strings.TrimSpace(req.Account)
	code := strings.TrimSpace(req.Code)
	if account == "" || code == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account, code and new_password are required"})
		return
	}
	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	var user store.User
	db := s.database()
	err := db.Where("username = ? OR email = ?", account, account).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "验证码无效或已过期"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var reset store.PasswordResetCode
	err = db.Where("user_id = ? AND used = ? AND expires_at > ?", user.ID, false, time.Now().UTC()).Order("id desc").First(&reset).Error
	if err != nil || !store.CheckPassword(code, reset.CodeHash) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "验证码无效或已过期"})
		return
	}
	passwordHash, err := store.HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]any{"password_hash": passwordHash, "enabled": true}).Error; err != nil {
			return err
		}
		if err := tx.Model(&reset).Update("used", true).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", user.ID).Delete(&store.UserSession{}).Error
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "密码已重设，请使用新密码登录"})
}

func (s *Server) userForRequest(r *http.Request) (store.User, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" || token == r.Header.Get("Authorization") {
		return store.User{}, false
	}

	var session store.UserSession
	err := s.database().Preload("User").Where("token = ? AND expires_at > ?", token, time.Now().UTC()).First(&session).Error
	if err != nil || !session.User.Enabled {
		return store.User{}, false
	}
	return session.User, true
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	user, ok := currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return false
	}
	if user.Role != store.UserRoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "administrator role required"})
		return false
	}
	return true
}

func canAccessPath(user store.User, path string) bool {
	if user.Role == store.UserRoleAdmin {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/auth/") {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/users") {
		return false
	}
	if strings.HasPrefix(path, "/api/v1/clusters") {
		return false
	}

	switch {
	case strings.Contains(path, "/configuration"), strings.Contains(path, "/logs"):
		return hasPermission(user, "system:read")
	case strings.Contains(path, "/storage"), strings.Contains(path, "/pools"), strings.Contains(path, "/block"), strings.Contains(path, "/filesystems"), strings.Contains(path, "/object"):
		return hasPermission(user, "storage:read")
	default:
		return hasPermission(user, "cluster:read")
	}
}

func hasPermission(user store.User, permission string) bool {
	var permissions []string
	if err := json.Unmarshal([]byte(user.Permissions), &permissions); err != nil {
		return false
	}
	for _, item := range permissions {
		if item == permission {
			return true
		}
	}
	return false
}

func buildUser(username, displayName, email, role string, permissions []string, password string, enabled *bool) (store.User, error) {
	username = strings.TrimSpace(username)
	displayName = strings.TrimSpace(displayName)
	if username == "" || displayName == "" || password == "" {
		return store.User{}, fmt.Errorf("username, display_name and password are required")
	}
	if len(password) < 8 {
		return store.User{}, fmt.Errorf("password must be at least 8 characters")
	}

	normalizedRole, err := normalizeRole(role)
	if err != nil {
		return store.User{}, err
	}
	passwordHash, err := store.HashPassword(password)
	if err != nil {
		return store.User{}, err
	}

	isEnabled := true
	if enabled != nil {
		isEnabled = *enabled
	}

	return store.User{
		Username:     username,
		DisplayName:  displayName,
		Email:        strings.TrimSpace(email),
		Role:         normalizedRole,
		Permissions:  permissionsJSON(permissions, normalizedRole),
		PasswordHash: passwordHash,
		Enabled:      isEnabled,
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
		permissions = defaultPermissions(role)
	}
	payload, err := json.Marshal(permissions)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func defaultPermissions(role string) []string {
	if role == store.UserRoleAdmin {
		return []string{"cluster:read", "storage:read", "system:read", "user:manage"}
	}
	return []string{"cluster:read", "storage:read"}
}

func toUserResponse(user store.User) userResponse {
	var permissions []string
	_ = json.Unmarshal([]byte(user.Permissions), &permissions)
	return userResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: permissions,
		Enabled:     user.Enabled,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func (s *Server) sendPasswordResetCode(user store.User, code string) error {
	cfg := s.currentConfig()
	if strings.TrimSpace(cfg.SMTP.Host) == "" {
		log.Printf("cephtower password reset code for %s <%s>: %s", user.Username, user.Email, code)
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
	subject := "CephTower 密码重置验证码"
	body := fmt.Sprintf("您的 CephTower 密码重置验证码是：%s\n\n验证码将在 %d 分钟后过期。", code, int(passwordResetTTL.Minutes()))
	message := strings.Join([]string{
		"From: " + from,
		"To: " + user.Email,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	if err := smtp.SendMail(addr, auth, from, []string{user.Email}, []byte(message)); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
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
