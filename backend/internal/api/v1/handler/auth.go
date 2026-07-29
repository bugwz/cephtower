package handler

import (
	"errors"
	"net/http"
	"time"

	"cephtower/backend/internal/config"
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

type bootstrapRunRequest struct {
	Database setupDatabaseRequest `json:"database"`
	User     createUserRequest    `json:"user"`
}

type setupDatabaseRequest struct {
	Engine string `json:"engine"`
	SQLite struct {
		Name string `json:"name"`
	} `json:"sqlite"`
	MySQL struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Params   string `json:"params"`
		TLS      string `json:"tls"`
	} `json:"mysql"`
}

func (h *Handler) BootstrapStatus(w http.ResponseWriter, r *http.Request) {
	if h.Setup == nil {
		WriteError(w, r, http.StatusServiceUnavailable, "setup_unavailable", "setup service is unavailable", true, nil)
		return
	}
	status, err := h.Setup.Status(r.Context())
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", "could not inspect initialization state", false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]bool{"required": status.Required})
}

func (h *Handler) BootstrapRun(w http.ResponseWriter, r *http.Request) {
	if h.Setup == nil {
		WriteError(w, r, http.StatusServiceUnavailable, "setup_unavailable", "setup service is unavailable", true, nil)
		return
	}
	var request bootstrapRunRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	row, err := h.Setup.Initialize(r.Context(), request.Database.toConfig(), authservice.CreateUserInput{Username: request.User.Username, DisplayName: request.User.DisplayName, Email: request.User.Email, Password: request.User.Password})
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

func (h *Handler) TestBootstrapDatabase(w http.ResponseWriter, r *http.Request) {
	if h.Setup == nil {
		WriteError(w, r, http.StatusServiceUnavailable, "setup_unavailable", "setup service is unavailable", true, nil)
		return
	}
	var request setupDatabaseRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if err := h.Setup.TestDatabase(r.Context(), request.toConfig()); err != nil {
		WriteError(w, r, http.StatusBadRequest, "database_unavailable", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]string{"status": "ok"})
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
	if h.Database == nil || h.Database() == nil {
		WriteError(w, r, http.StatusServiceUnavailable, "not_ready", "database is unavailable", true, nil)
		return
	}
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

func (request setupDatabaseRequest) toConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Engine: request.Engine,
		SQLite: config.SQLiteConfig{
			Name: request.SQLite.Name,
		},
		MySQL: config.MySQLConfig{
			Host:     request.MySQL.Host,
			Port:     request.MySQL.Port,
			Username: request.MySQL.Username,
			Password: request.MySQL.Password,
			Database: request.MySQL.Database,
			Params:   request.MySQL.Params,
			TLS:      request.MySQL.TLS,
		},
	}
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
