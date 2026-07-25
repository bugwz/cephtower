package handler

import (
	"errors"
	"net/http"

	"cephtower/backend/internal/config"
	setupservice "cephtower/backend/internal/service/setup"
)

type SetupDatabaseResponse struct {
	Engine string `json:"engine"`
	SQLite struct {
		Path string `json:"path"`
	} `json:"sqlite"`
	MySQL struct {
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Username    string `json:"username"`
		Password    string `json:"password"`
		PasswordSet bool   `json:"password_set"`
		Database    string `json:"database"`
		Params      string `json:"params"`
		TLS         string `json:"tls"`
	} `json:"mysql"`
}
type SetupDatabaseRequest struct {
	Engine string `json:"engine"`
	SQLite struct {
		Path string `json:"path"`
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

type SetupInitializeRequest struct {
	Database SetupDatabaseRequest `json:"database"`
	Admin    struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"admin"`
}

func (h *Handler) TestSetupDatabase(w http.ResponseWriter, r *http.Request) {
	var request SetupDatabaseRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err := h.setup.TestDatabase(r.Context(), setupDatabaseConfigFromRequest(request))
	switch {
	case errors.Is(err, setupservice.ErrAlreadyInitialized):
		writeError(w, http.StatusConflict, err)
	case err != nil:
		writeError(w, http.StatusBadRequest, err)
	default:
		writeJSON(w, http.StatusOK, MessageResponse{Message: "database connection succeeded"})
	}
}

func (h *Handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	initialized, database, err := h.setup.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	response := map[string]any{"initialized": initialized}
	if !initialized {
		response["database"] = setupDatabaseFromConfig(database)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) InitializeSetup(w http.ResponseWriter, r *http.Request) {
	var request SetupInitializeRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	input := setupservice.Input{Database: setupDatabaseConfigFromRequest(request.Database), Username: request.Admin.Username, Email: request.Admin.Email, Password: request.Admin.Password}
	err := h.setup.Initialize(r.Context(), input)
	switch {
	case errors.Is(err, setupservice.ErrAlreadyInitialized), errors.Is(err, setupservice.ErrTargetInitialized):
		writeError(w, http.StatusConflict, err)
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "system initialized"})
}

func setupDatabaseConfigFromRequest(request SetupDatabaseRequest) config.DatabaseConfig {
	return config.DatabaseConfig{
		Engine: request.Engine,
		SQLite: config.SQLiteConfig{Path: request.SQLite.Path},
		MySQL: config.MySQLConfig{
			Host: request.MySQL.Host, Port: request.MySQL.Port,
			Username: request.MySQL.Username, Password: request.MySQL.Password,
			Database: request.MySQL.Database, Params: request.MySQL.Params, TLS: request.MySQL.TLS,
		},
	}
}

func setupDatabaseFromConfig(cfg config.DatabaseConfig) SetupDatabaseResponse {
	var result SetupDatabaseResponse
	result.Engine = cfg.Engine
	result.SQLite.Path = cfg.SQLite.Path
	result.MySQL.Host = cfg.MySQL.Host
	result.MySQL.Port = cfg.MySQL.Port
	result.MySQL.Username = cfg.MySQL.Username
	result.MySQL.Password = cfg.MySQL.Password
	result.MySQL.PasswordSet = cfg.MySQL.Password != ""
	result.MySQL.Database = cfg.MySQL.Database
	result.MySQL.Params = cfg.MySQL.Params
	result.MySQL.TLS = cfg.MySQL.TLS
	return result
}

type SetupStatusResponse struct {
	Initialized bool                   `json:"initialized"`
	Database    *SetupDatabaseResponse `json:"database,omitempty"`
}
