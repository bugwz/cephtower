package v1

import (
	"errors"
	"net/http"
	"strings"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/service/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type setupDatabaseResponse struct {
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

type setupInitializeRequest struct {
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

func (api *API) SetupStatus(w http.ResponseWriter, _ *http.Request) {
	db := api.database()
	initialized, err := hasUsers(db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := map[string]any{"initialized": initialized}
	if !initialized {
		response["database"] = setupDatabaseFromConfig(api.currentConfig().Database)
	}
	writeJSON(w, http.StatusOK, response)
}

func (api *API) InitializeSetup(w http.ResponseWriter, r *http.Request) {
	var req setupInitializeRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	username := strings.TrimSpace(req.Admin.Username)
	email := strings.TrimSpace(req.Admin.Email)
	if username == "" || email == "" || req.Admin.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "admin username, email and password are required"})
		return
	}
	if len(req.Admin.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	currentDB := api.database()
	currentCfg := api.currentConfig()
	currentHasUsers, err := hasUsers(currentDB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if currentHasUsers {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "system has already been initialized"})
		return
	}

	databaseCfg, err := normalizeSetupDatabase(req, currentCfg.Database)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	newDB, err := store.Open(databaseCfg, currentCfg.Server.Dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	targetHasUsers, err := hasUsers(newDB)
	if err != nil {
		_ = store.Close(newDB)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if targetHasUsers {
		_ = store.Close(newDB)
		writeJSON(w, http.StatusConflict, map[string]string{"error": "selected database has already been initialized"})
		return
	}

	admin, err := buildSetupAdmin(username, email, req.Admin.Password)
	if err != nil {
		_ = store.Close(newDB)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := newDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		if err := ceph.EnsureDefaultSystemSettings(r.Context(), tx); err != nil {
			return err
		}
		return config.SaveDatabase(currentCfg.Path, databaseCfg)
	}); err != nil {
		_ = store.Close(newDB)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	currentCfg.Database = databaseCfg
	previousDB := api.replaceDatabase(currentCfg, newDB)

	if previousDB != nil && previousDB != newDB {
		_ = store.Close(previousDB)
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "system initialized"})
}

func hasUsers(db *gorm.DB) (bool, error) {
	var count int64
	if err := db.Model(&store.User{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func normalizeSetupDatabase(req setupInitializeRequest, current config.DatabaseConfig) (config.DatabaseConfig, error) {
	cfg := config.DatabaseConfig{
		Engine: req.Database.Engine,
		SQLite: config.SQLiteConfig{
			Path: req.Database.SQLite.Path,
		},
		MySQL: config.MySQLConfig{
			Host:     req.Database.MySQL.Host,
			Port:     req.Database.MySQL.Port,
			Username: req.Database.MySQL.Username,
			Password: req.Database.MySQL.Password,
			Database: req.Database.MySQL.Database,
			Params:   req.Database.MySQL.Params,
		},
	}
	if cfg.Engine == "mysql" && strings.TrimSpace(cfg.MySQL.Password) == "" {
		return config.DatabaseConfig{}, errors.New("mysql password is required")
	}
	return config.NormalizeDatabaseConfig(cfg)
}

func buildSetupAdmin(username, email, password string) (store.User, error) {
	passwordHash, err := store.HashPassword(password)
	if err != nil {
		return store.User{}, err
	}
	return store.User{
		Username:     username,
		DisplayName:  username,
		Email:        email,
		Role:         store.UserRoleAdmin,
		Permissions:  permissionsJSON(nil, store.UserRoleAdmin),
		PasswordHash: passwordHash,
		Enabled:      true,
	}, nil
}

func setupDatabaseFromConfig(cfg config.DatabaseConfig) setupDatabaseResponse {
	var response setupDatabaseResponse
	response.Engine = cfg.Engine
	response.SQLite.Path = cfg.SQLite.Path
	response.MySQL.Host = cfg.MySQL.Host
	response.MySQL.Port = cfg.MySQL.Port
	response.MySQL.Username = cfg.MySQL.Username
	response.MySQL.Password = cfg.MySQL.Password
	response.MySQL.PasswordSet = cfg.MySQL.Password != ""
	response.MySQL.Database = cfg.MySQL.Database
	response.MySQL.Params = cfg.MySQL.Params
	return response
}
