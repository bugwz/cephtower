package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	appconfig "cephtower/backend/internal/config"
	"cephtower/backend/internal/service/ceph"
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

func (api *API) ListSystemSettings(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := ceph.EnsureDefaultSystemSettings(r.Context(), api.database()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	var settings []store.Setting
	query := api.database().WithContext(r.Context()).Order("`key` asc")
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

func (api *API) UpdateSystemSetting(w http.ResponseWriter, r *http.Request) {
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
	if strings.HasPrefix(key, ceph.DataFetchSettingPrefix) {
		var config ceph.DataFetchConfig
		if err := json.Unmarshal([]byte(req.Value), &config); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data fetch setting value must be valid JSON"})
			return
		}
		if config.Module == "" {
			config.Module = strings.TrimPrefix(key, ceph.DataFetchSettingPrefix)
		}
		if config.Module != strings.TrimPrefix(key, ceph.DataFetchSettingPrefix) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data fetch setting module must match setting key"})
			return
		}
		if err := validateDataFetchConfig(config); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		ceph.NormalizeDataFetchConfig(&config)
		data, err := json.Marshal(config)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		req.Value = string(data)
	}
	setting := store.Setting{Key: key, Value: req.Value}
	if err := api.database().WithContext(r.Context()).
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

func (api *API) ResetSystemConfigDefaults(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := api.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`key` LIKE ?", ceph.DataFetchSettingPrefix+"%").Delete(&store.Setting{}).Error; err != nil {
			return err
		}
		return ceph.EnsureDefaultSystemSettings(r.Context(), tx)
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "系统配置已恢复默认"})
}

func (api *API) RunDataFetchModuleNow(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	module := strings.TrimSpace(r.PathValue("module"))
	config, ok, err := dataFetchConfigByModule(r.Context(), api.database(), module)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "data fetch module setting not found"})
		return
	}
	var clusters []store.CephCluster
	if err := api.database().WithContext(r.Context()).Order("id asc").Find(&clusters).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	ctx := context.WithoutCancel(r.Context())
	service := ceph.NewService(api.database, appconfig.ResolveRuntimeDir(api.currentConfig()))
	go func() {
		for _, cluster := range clusters {
			_ = service.RunDataFetchConfig(ctx, cluster.ID, config)
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "数据获取任务已启动"})
}

func (api *API) ListDataFetchRuns(w http.ResponseWriter, r *http.Request) {
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
	query := api.database().WithContext(r.Context()).Order("started_at desc").Limit(limit)
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

func dataFetchConfigByModule(ctx context.Context, db *gorm.DB, module string) (ceph.DataFetchConfig, bool, error) {
	var setting store.Setting
	err := db.WithContext(ctx).Where("`key` = ?", ceph.DataFetchSettingKey(module)).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return ceph.DataFetchConfig{}, false, nil
	}
	if err != nil {
		return ceph.DataFetchConfig{}, false, err
	}
	var config ceph.DataFetchConfig
	if err := json.Unmarshal([]byte(setting.Value), &config); err != nil {
		return ceph.DataFetchConfig{}, false, err
	}
	if config.Module == "" {
		config.Module = module
	}
	ceph.NormalizeDataFetchConfig(&config)
	return config, true, nil
}

func validateDataFetchConfig(config ceph.DataFetchConfig) error {
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
	if config.FetchSource != "command" && config.FetchSource != "dashboard" && config.FetchSource != "mixed" {
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
