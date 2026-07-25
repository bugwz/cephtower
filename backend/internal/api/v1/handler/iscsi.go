package handler

import "net/http"

// iSCSI endpoints.
func (h *Handler) GetISCSISettings(w http.ResponseWriter, r *http.Request) {
	h.GetSettingGroup(w, r)
}

func (h *Handler) UpdateISCSISettings(w http.ResponseWriter, r *http.Request) {
	h.UpdateSettingGroup(w, r)
}

func (h *Handler) GetISCSIDiscoveryAuth(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/discoveryauth")
}

func (h *Handler) UpdateISCSIDiscoveryAuth(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/discoveryauth")
}

func (h *Handler) ListISCSITargets(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/target")
}

func (h *Handler) CreateISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/target")
}

func (h *Handler) GetISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/target/{iqn}")
}

func (h *Handler) UpdateISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/target/{iqn}")
}

func (h *Handler) DeleteISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/iscsi/target/{iqn}")
}

type ISCSISettingsRequest = RawJSONRequest
type ISCSISettingsResponse = RawJSONResponse
type ISCSIDiscoveryAuthRequest = RawJSONRequest
type ISCSIDiscoveryAuthResponse = RawJSONResponse
type ISCSITargetRequest = RawJSONRequest
type ISCSITargetResponse = RawJSONResponse
