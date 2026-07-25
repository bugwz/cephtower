package handler

import "net/http"

// Prometheus endpoints.
func (h *Handler) GetPrometheusSettings(w http.ResponseWriter, r *http.Request) {
	h.GetSettingGroup(w, r)
}

func (h *Handler) UpdatePrometheusSettings(w http.ResponseWriter, r *http.Request) {
	h.UpdateSettingGroup(w, r)
}

func (h *Handler) GetPrometheus(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus")
}

func (h *Handler) ListPrometheusAlertGroups(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/alertgroup")
}

func (h *Handler) GetPrometheusData(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/data")
}

func (h *Handler) ListPrometheusNotifications(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/notifications")
}

func (h *Handler) QueryPrometheus(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/prometheus_query_data")
}

func (h *Handler) ListPrometheusRules(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/rules")
}

func (h *Handler) ListPrometheusSilences(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/silences")
}

func (h *Handler) CreatePrometheusSilence(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/silence")
}

func (h *Handler) DeletePrometheusSilence(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/prometheus/silence/{id}")
}

type PrometheusSettingsRequest = RawJSONRequest
type PrometheusSettingsResponse = RawJSONResponse
type PrometheusResponse = RawJSONResponse
type PrometheusSilenceRequest = RawJSONRequest
type PrometheusSilenceResponse = RawJSONResponse
