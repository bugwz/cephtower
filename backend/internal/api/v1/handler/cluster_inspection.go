package handler

import "net/http"

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Channel   string `json:"channel,omitempty"`
		Level     string `json:"level,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "log.list", "log", request.Channel, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.Logs(r.Context(), request.ClusterID, request.Channel, request.Level, request.Limit)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}
func (h *Handler) GetConfigurationOption(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Name      string `json:"name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "config_option.get", "config_option", request.Name, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.ConfigurationOption(r.Context(), request.ClusterID, request.Name)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetOSDInspection(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		OSDID     string `json:"osd_id"`
		Section   string `json:"section"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "osd.inspect", "osd", request.OSDID, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.OSDInspection(r.Context(), request.ClusterID, request.OSDID, request.Section)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetSnapshotScheduleStatus(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		FS        string `json:"fs"`
		Path      string `json:"path"`
		Subvol    string `json:"subvol,omitempty"`
		Group     string `json:"group,omitempty"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "snapshot_schedule.status", "snapshot_schedule", request.Path, "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	rows, err := h.Inspection.SnapshotSchedules(r.Context(), request.ClusterID, request.FS, request.Path, request.Subvol, request.Group)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", map[string]any{"items": rows})
}
