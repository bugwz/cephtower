package handler

import (
	"net/http"
	"time"
)

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
