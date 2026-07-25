package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	settingsservice "cephtower/backend/internal/service/settings"
	"cephtower/backend/internal/store"
)

type SystemSettingResponse struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
type UpdateSystemSettingRequest struct {
	Value string `json:"value"`
}
type DataFetchRunResponse struct {
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

func (h *Handler) ListSystemSettings(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	items, err := h.systemSettings.List(r.Context(), r.URL.Query().Get("prefix"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	result := make([]SystemSettingResponse, 0, len(items))
	for _, item := range items {
		result = append(result, SystemSettingResponse{Key: item.Key, Value: item.Value, UpdatedAt: item.UpdatedAt})
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateSystemSetting(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var request UpdateSystemSettingRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.systemSettings.Update(r.Context(), r.PathValue("key"), request.Value)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, SystemSettingResponse{Key: item.Key, Value: item.Value, UpdatedAt: item.UpdatedAt})
}

func (h *Handler) ResetSystemConfigDefaults(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.systemSettings.Reset(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "系统配置已恢复默认"})
}

func (h *Handler) RunDataFetchModuleNow(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	err := h.systemSettings.RunModule(r.Context(), strings.TrimSpace(r.PathValue("module")))
	switch {
	case errors.Is(err, settingsservice.ErrDataFetchConfigNotFound):
		writeError(w, http.StatusNotFound, err)
		return
	case err != nil:
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "数据获取任务已启动"})
}

func (h *Handler) ListDataFetchRuns(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	items, err := h.systemSettings.ListRuns(r.Context(), r.URL.Query().Get("cluster_id"), r.URL.Query().Get("module"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	result := make([]DataFetchRunResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toDataFetchRunResponse(item))
	}
	writeJSON(w, http.StatusOK, result)
}

func toDataFetchRunResponse(run store.CephDataFetchRun) DataFetchRunResponse {
	return DataFetchRunResponse{ID: run.ID, ClusterID: run.ClusterID, Module: run.Module, Status: run.Status, Source: run.Source, StartedAt: run.StartedAt, FinishedAt: run.FinishedAt, DurationMS: run.DurationMS, RecordsUpserted: run.RecordsUpserted, RecordsDeleted: run.RecordsDeleted, Error: run.Error, CreatedAt: run.CreatedAt}
}

type ListSystemSettingsRequest struct {
	Prefix string `json:"prefix"`
}
type ListDataFetchRunsRequest struct {
	ClusterID string `json:"cluster_id"`
	Module    string `json:"module"`
	Limit     int    `json:"limit"`
}
