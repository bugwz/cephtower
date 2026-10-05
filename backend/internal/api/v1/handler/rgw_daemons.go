package handler

import (
	"net/http"
	"time"
)

func (h *Handler) GetRGWDaemonPerf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var input struct {
		ClusterID    uint64 `json:"cluster_id"`
		ServiceMapID string `json:"service_map_id"`
	}
	if !DecodeStrict(w, r, &input) {
		return
	}
	annotateAudit(r, "rgw.daemon.perf", "rgw_daemon", input.ServiceMapID, "", &input.ClusterID)
	if h.External == nil {
		WriteError(w, r, 501, "capability_unavailable", "monitoring reader is unavailable", false, nil)
		return
	}
	result, err := h.External.RGWPerf(r.Context(), input.ClusterID, input.ServiceMapID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetRGWDaemonStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var input struct {
		ClusterID    uint64 `json:"cluster_id"`
		ServiceMapID string `json:"service_map_id"`
	}
	if !DecodeStrict(w, r, &input) {
		return
	}
	annotateAudit(r, "rgw.daemon.status", "rgw_daemon", input.ServiceMapID, "", &input.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RGWDaemonStatus(r.Context(), input.ClusterID, input.ServiceMapID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) ListRGWDaemons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var input struct {
		ClusterID uint64 `json:"cluster_id"`
	}
	if !DecodeStrict(w, r, &input) {
		return
	}
	if input.ClusterID == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	annotateAudit(r, "rgw.daemons.read", "rgw_daemon", "", "", &input.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	rows, err := h.Inspection.RGWDaemons(r.Context(), input.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": rows, "observed_at": time.Now().UTC(), "source": "service_map"})
}
