package handler

import "net/http"

func (h *Handler) GetRBDMirrorScheduleStatus(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rbd.mirror.schedule.status", "rbd_mirroring", "schedule/status", "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RBDMirrorScheduleStatus(r.Context(), request.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetRBDMirrorSchedules(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rbd.mirror.schedules", "rbd_mirroring", "schedules", "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RBDMirrorSchedules(r.Context(), request.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}
