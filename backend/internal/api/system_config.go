package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type systemSettingResponse struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type updateSystemSettingRequest struct {
	Value string `json:"value"`
}

type dataFetchRunResponse struct {
	ID              uint       `json:"id"`
	ClusterID       uint       `json:"cluster_id"`
	Module          string     `json:"module"`
	Status          string     `json:"status"`
	Source          string     `json:"source"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	DurationMS      int        `json:"duration_ms"`
	RecordsUpserted int        `json:"records_upserted"`
	RecordsDeleted  int        `json:"records_deleted"`
	Error           string     `json:"error"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (s *Server) registerSystemConfigRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/system/config/settings", s.listSystemSettings)
	mux.HandleFunc("PUT /api/v1/system/config/settings/{key}", s.updateSystemSetting)
	mux.HandleFunc("POST /api/v1/system/config/defaults/reset", s.resetSystemConfigDefaults)
	mux.HandleFunc("POST /api/v1/system/config/data-fetch/{module}/run", s.runDataFetchModuleNow)
	mux.HandleFunc("GET /api/v1/system/config/data-fetch/runs", s.listDataFetchRuns)
}

func (s *Server) listSystemSettings(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := ensureDefaultSystemSettings(r.Context(), s.database()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	var settings []store.Setting
	query := s.database().WithContext(r.Context()).Order("`key` asc")
	if prefix := strings.TrimSpace(r.URL.Query().Get("prefix")); prefix != "" {
		query = query.Where("`key` LIKE ?", prefix+"%")
	}
	if err := query.Find(&settings).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	response := make([]systemSettingResponse, 0, len(settings))
	for _, setting := range settings {
		response = append(response, systemSettingResponse{
			Key:       setting.Key,
			Value:     setting.Value,
			UpdatedAt: setting.UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) updateSystemSetting(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "setting key is required"})
		return
	}
	var req updateSystemSettingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.HasPrefix(key, dataFetchSettingPrefix) {
		var config dataFetchConfig
		if err := json.Unmarshal([]byte(req.Value), &config); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data fetch setting value must be valid JSON"})
			return
		}
		if config.Module == "" {
			config.Module = strings.TrimPrefix(key, dataFetchSettingPrefix)
		}
		if config.Module != strings.TrimPrefix(key, dataFetchSettingPrefix) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data fetch setting module must match setting key"})
			return
		}
		if err := validateDataFetchConfig(config); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		normalizeDataFetchConfig(&config)
		req.Value = mustJSON(config)
	}
	setting := store.Setting{Key: key, Value: req.Value}
	if err := s.database().WithContext(r.Context()).
		Where("`key` = ?", key).
		Assign(store.Setting{Value: req.Value}).
		FirstOrCreate(&setting).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, systemSettingResponse{
		Key:       setting.Key,
		Value:     setting.Value,
		UpdatedAt: setting.UpdatedAt,
	})
}

func (s *Server) resetSystemConfigDefaults(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := s.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`key` LIKE ?", dataFetchSettingPrefix+"%").Delete(&store.Setting{}).Error; err != nil {
			return err
		}
		return ensureDefaultSystemSettings(r.Context(), tx)
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "系统配置已恢复默认"})
}

func (s *Server) runDataFetchModuleNow(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	module := strings.TrimSpace(r.PathValue("module"))
	config, ok, err := dataFetchConfigByModule(r.Context(), s.database(), module)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "data fetch module setting not found"})
		return
	}
	var clusters []store.CephCluster
	if err := s.database().WithContext(r.Context()).Order("id asc").Find(&clusters).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	ctx := context.WithoutCancel(r.Context())
	go func() {
		for _, cluster := range clusters {
			_ = s.runDataFetchConfig(ctx, cluster.ID, config)
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "数据获取任务已启动"})
}

func (s *Server) listDataFetchRuns(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}
	var runs []store.CephDataFetchRun
	query := s.database().WithContext(r.Context()).Order("started_at desc").Limit(limit)
	if clusterID := strings.TrimSpace(r.URL.Query().Get("cluster_id")); clusterID != "" {
		query = query.Where("cluster_id = ?", clusterID)
	}
	if module := strings.TrimSpace(r.URL.Query().Get("module")); module != "" {
		query = query.Where("module = ?", module)
	}
	if err := query.Find(&runs).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	response := make([]dataFetchRunResponse, 0, len(runs))
	for _, run := range runs {
		response = append(response, toDataFetchRunResponse(run))
	}
	writeJSON(w, http.StatusOK, response)
}

func dataFetchConfigByModule(ctx context.Context, db *gorm.DB, module string) (dataFetchConfig, bool, error) {
	var setting store.Setting
	err := db.WithContext(ctx).Where("`key` = ?", dataFetchSettingKey(module)).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return dataFetchConfig{}, false, nil
	}
	if err != nil {
		return dataFetchConfig{}, false, err
	}
	var config dataFetchConfig
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		return dataFetchConfig{}, false, err
	}
	if config.Module == "" {
		config.Module = module
	}
	normalizeDataFetchConfig(&config)
	return config, true, nil
}

func validateDataFetchConfig(config dataFetchConfig) error {
	if strings.TrimSpace(config.Module) == "" {
		return errInvalidSystemConfig("module is required")
	}
	if config.IntervalSeconds < 10 {
		return errInvalidSystemConfig("interval_seconds must be at least 10")
	}
	if config.TimeoutSeconds < 1 {
		return errInvalidSystemConfig("timeout_seconds must be positive")
	}
	if config.JitterSeconds < 0 {
		return errInvalidSystemConfig("jitter_seconds cannot be negative")
	}
	if config.FetchSource != fetchSourceCommand && config.FetchSource != fetchSourceDashboard && config.FetchSource != "mixed" {
		return errInvalidSystemConfig("fetch_source must be command, dashboard or mixed")
	}
	if config.MaxRetries < 0 {
		return errInvalidSystemConfig("max_retries cannot be negative")
	}
	if config.RetryBackoffSeconds < 0 {
		return errInvalidSystemConfig("retry_backoff_seconds cannot be negative")
	}
	return nil
}

type errInvalidSystemConfig string

func (e errInvalidSystemConfig) Error() string {
	return string(e)
}

func toDataFetchRunResponse(run store.CephDataFetchRun) dataFetchRunResponse {
	return dataFetchRunResponse{
		ID:              run.ID,
		ClusterID:       run.ClusterID,
		Module:          run.Module,
		Status:          run.Status,
		Source:          run.Source,
		StartedAt:       run.StartedAt,
		FinishedAt:      run.FinishedAt,
		DurationMS:      run.DurationMS,
		RecordsUpserted: run.RecordsUpserted,
		RecordsDeleted:  run.RecordsDeleted,
		Error:           run.Error,
		CreatedAt:       run.CreatedAt,
	}
}
