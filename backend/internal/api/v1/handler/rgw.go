package handler

import "net/http"

// RGW endpoints.
func (h *Handler) GetRGWSettings(w http.ResponseWriter, r *http.Request) {
	h.GetSettingGroup(w, r)
}

func (h *Handler) UpdateRGWSettings(w http.ResponseWriter, r *http.Request) {
	h.UpdateSettingGroup(w, r)
}

func (h *Handler) GetRGWStatus(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/rgw/daemon")
}

func (h *Handler) ValidateRGWConfig(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/rgw/daemon")
}

type RGWSettingsRequest = RawJSONRequest
type RGWSettingsResponse = RawJSONResponse
type RGWStatusResponse = RawJSONResponse
type RGWValidateRequest = RawJSONRequest
type RGWValidateResponse = RawJSONResponse
