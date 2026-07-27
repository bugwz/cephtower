package handler

import (
	"errors"
	"net/http"
	"time"

	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/store"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
}

func (h *Handler) BootstrapStatus(w http.ResponseWriter, r *http.Request) {
	required, err := h.Auth.BootstrapRequired(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", "could not inspect initialization state", false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]bool{"required": required})
}

func (h *Handler) BootstrapAdmin(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	row, err := h.Auth.CreateInitialAdmin(r.Context(), authservice.CreateUserInput{Username: request.Username, DisplayName: request.DisplayName, Email: request.Email, Password: request.Password})
	if err != nil {
		status, code := http.StatusBadRequest, "invalid_request"
		if errors.Is(err, authservice.ErrAlreadyInitialized) {
			status, code = http.StatusConflict, "already_initialized"
		}
		WriteError(w, r, status, code, err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusCreated, "success", toUserDTO(row))
}

type userDTO struct {
	ID          uint64     `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       *string    `json:"email"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	result, err := h.Auth.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, authservice.ErrUserDisabled) {
			status = http.StatusForbidden
		}
		WriteError(w, r, status, "invalid_credentials", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]any{"token": result.Token, "expires_at": result.ExpiresAt, "user": toUserDTO(result.User)})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Auth.ListUsers(r.Context())
	if err != nil {
		WriteError(w, r, 500, "store_error", err.Error(), false, nil)
		return
	}
	items := make([]userDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUserDTO(row))
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]string{"request_id": RequestID(r)}})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	row, err := h.Auth.CreateUser(r.Context(), authservice.CreateUserInput{Username: request.Username, DisplayName: request.DisplayName, Email: request.Email, Password: request.Password, Role: request.Role})
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 201, "success", toUserDTO(row))
}

func toUserDTO(user store.User) userDTO {
	return userDTO{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Email: user.Email, Status: user.Status, LastLoginAt: user.LastLoginAt, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
}
