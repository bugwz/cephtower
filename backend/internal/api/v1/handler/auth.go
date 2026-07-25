package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/store"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

type UserResponse struct {
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

type CreateUserRequest struct {
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Password    string   `json:"password"`
	Enabled     *bool    `json:"enabled"`
}

type UpdateUserRequest struct {
	DisplayName *string  `json:"display_name"`
	Email       *string  `json:"email"`
	Role        *string  `json:"role"`
	Permissions []string `json:"permissions"`
	Password    *string  `json:"password"`
	Enabled     *bool    `json:"enabled"`
}

type PasswordResetRequest struct {
	Account string `json:"account"`
}

type PasswordResetConfirmRequest struct {
	Account     string `json:"account"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, authservice.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, LoginResponse{
		Token: result.Token, ExpiresAt: result.ExpiresAt, User: toUserResponse(result.User),
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	users, err := h.auth.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, toUserResponse(user))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var req CreateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := h.auth.CreateUser(r.Context(), authservice.CreateUserInput{
		Username: req.Username, DisplayName: req.DisplayName, Email: req.Email, Role: req.Role,
		Permissions: req.Permissions, Password: req.Password, Enabled: req.Enabled,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var req UpdateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := h.auth.UpdateUser(r.Context(), uint(id), authservice.UpdateUserInput{
		DisplayName: req.DisplayName, Email: req.Email, Role: req.Role, Permissions: req.Permissions,
		Password: req.Password, Enabled: req.Enabled,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, authservice.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.auth.RequestPasswordReset(r.Context(), req.Account); err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, authservice.ErrEmailRequired) || strings.TrimSpace(req.Account) == "" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "验证码已发送，请查收邮箱"})
}

func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetConfirmRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.auth.ConfirmPasswordReset(r.Context(), req.Account, req.Code, req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "密码已重设，请使用新密码登录"})
}

func toUserResponse(user store.User) UserResponse {
	var permissions []string
	_ = json.Unmarshal([]byte(user.Permissions), &permissions)
	return UserResponse{
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

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}
