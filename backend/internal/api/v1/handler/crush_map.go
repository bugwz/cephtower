package handler

import "net/http"

func (h *Handler) GetCrushMap(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "crush.map", "crush_rule", "map", "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.CrushMap(r.Context(), request.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, 200, "success", result)
}
