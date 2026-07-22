package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/service/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

func TestSetupInitializeIsOnlyAvailableBeforeUsersExist(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	dbPath := filepath.Join(tempDir, "cephtower.db")
	configData := []byte(`database:
  engine: sqlite
  sqlite:
    path: ` + dbPath + `
  mysql:
    password: configured-secret
`)
	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	db, err := store.Open(cfg.Database)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	currentCfg := cfg
	currentDB := db
	api := NewAPI(nil, Dependencies{
		CurrentConfig: func() config.Config {
			return currentCfg
		},
		Database: func() *gorm.DB {
			return currentDB
		},
		ReplaceDatabase: func(cfg config.Config, db *gorm.DB) *gorm.DB {
			previous := currentDB
			currentCfg = cfg
			currentDB = db
			return previous
		},
	})
	defer func() {
		if currentDB != nil {
			if err := store.Close(currentDB); err != nil {
				t.Fatalf("Close() returned error: %v", err)
			}
		}
	}()

	recorder := httptest.NewRecorder()
	api.SetupStatus(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("setup status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var status struct {
		Initialized bool           `json:"initialized"`
		Database    map[string]any `json:"database"`
	}
	if err := decodeAPIResponseData(recorder, &status); err != nil {
		t.Fatalf("decode setup status: %v", err)
	}
	if status.Initialized || status.Database == nil {
		t.Fatalf("status = %#v, want uninitialized response with database config", status)
	}
	mysqlConfig, ok := status.Database["mysql"].(map[string]any)
	if !ok {
		t.Fatalf("status database mysql = %#v, want object", status.Database["mysql"])
	}
	if mysqlConfig["password"] != "configured-secret" {
		t.Fatalf("status mysql password = %#v, want configured-secret", mysqlConfig["password"])
	}

	payload := []byte(`{
		"database": {
			"engine": "sqlite",
			"sqlite": {"path": "` + dbPath + `"},
			"mysql": {
				"host": "127.0.0.1",
				"port": 3306,
				"username": "cephtower",
				"database": "cephtower",
				"params": "charset=utf8mb4&parseTime=True&loc=Local"
			}
		},
		"admin": {
			"username": "admin",
			"email": "admin@example.com",
			"password": "ChangeMe123!"
		}
	}`)
	recorder = httptest.NewRecorder()
	api.InitializeSetup(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/setup/initialize", bytes.NewReader(payload)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("initialize = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	var admin store.User
	if err := currentDB.Where("username = ?", "admin").First(&admin).Error; err != nil {
		t.Fatalf("admin user was not created: %v", err)
	}
	if admin.Role != store.UserRoleAdmin || !admin.Enabled {
		t.Fatalf("admin = %#v, want enabled administrator", admin)
	}

	recorder = httptest.NewRecorder()
	api.InitializeSetup(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/setup/initialize", bytes.NewReader(payload)))
	if recorder.Code != http.StatusOK {
		t.Fatalf("second initialize = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	envelope, err := decodeAPIResponseEnvelope(recorder)
	if err != nil {
		t.Fatalf("decode second initialize response: %v", err)
	}
	if envelope.Code != http.StatusConflict {
		t.Fatalf("second initialize code = %d, want 409: %s", envelope.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	api.SetupStatus(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("setup status after init = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	status = struct {
		Initialized bool           `json:"initialized"`
		Database    map[string]any `json:"database"`
	}{}
	if err := decodeAPIResponseData(recorder, &status); err != nil {
		t.Fatalf("decode setup status after init: %v", err)
	}
	if !status.Initialized || status.Database != nil {
		t.Fatalf("status after init = %#v, want initialized response without database config", status)
	}

	var settingCount int64
	if err := currentDB.Model(&store.Setting{}).Where("`key` LIKE ?", ceph.DataFetchSettingPrefix+"%").Count(&settingCount).Error; err != nil {
		t.Fatalf("count default data fetch settings: %v", err)
	}
	if settingCount == 0 {
		t.Fatal("default data fetch settings were not initialized")
	}
}

func decodeAPIResponseData(recorder *httptest.ResponseRecorder, out any) error {
	response, err := decodeAPIResponseEnvelope(recorder)
	if err != nil {
		return err
	}
	if response.Code != 0 {
		return errors.New(response.Message)
	}
	return json.Unmarshal(response.Data, out)
}

type apiResponseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeAPIResponseEnvelope(recorder *httptest.ResponseRecorder) (apiResponseEnvelope, error) {
	var response apiResponseEnvelope
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	return response, err
}
