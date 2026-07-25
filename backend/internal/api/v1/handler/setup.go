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
	} `json:"mysql"`
}
type SetupInitializeRequest struct {
	Database struct {
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
		} `json:"mysql"`
	} `json:"database"`
	Admin struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"admin"`
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
	input := setupservice.Input{Database: config.DatabaseConfig{Engine: request.Database.Engine, SQLite: config.SQLiteConfig{Path: request.Database.SQLite.Path}, MySQL: config.MySQLConfig{Host: request.Database.MySQL.Host, Port: request.Database.MySQL.Port, Username: request.Database.MySQL.Username, Password: request.Database.MySQL.Password, Database: request.Database.MySQL.Database, Params: request.Database.MySQL.Params}}, Username: request.Admin.Username, Email: request.Admin.Email, Password: request.Admin.Password}
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
	return result
}

type SetupStatusResponse struct {
	Initialized bool                   `json:"initialized"`
	Database    *SetupDatabaseResponse `json:"database,omitempty"`
}
