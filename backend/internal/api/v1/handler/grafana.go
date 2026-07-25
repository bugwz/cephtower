package handler

import "net/http"

// Grafana endpoints.
func (h *Handler) GetGrafanaSettings(w http.ResponseWriter, r *http.Request) {
	h.GetSettingGroup(w, r)
}

func (h *Handler) UpdateGrafanaSettings(w http.ResponseWriter, r *http.Request) {
	h.UpdateSettingGroup(w, r)
}

func (h *Handler) GetGrafanaURL(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/grafana/url")
}

func (h *Handler) ValidateGrafana(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/grafana/validation/{params}")
}

func (h *Handler) SyncGrafanaDashboards(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/grafana/dashboards")
}

type GrafanaSettingsRequest = RawJSONRequest
type GrafanaSettingsResponse = RawJSONResponse
type GrafanaURLResponse = RawJSONResponse
type GrafanaValidationResponse = RawJSONResponse
type GrafanaDashboardSyncResponse = RawJSONResponse
