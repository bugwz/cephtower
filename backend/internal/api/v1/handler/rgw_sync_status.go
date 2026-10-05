package handler

import "net/http"

func (h *Handler) ReadRGWSyncStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		ZoneID    string `json:"zone_id"`
		Name      string `json:"name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rgw_zone.sync.status", "rgw_zone", request.ZoneID, "low", &request.ClusterID)
	if h.Mutations == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "Zone sync status reads are unavailable", false, nil)
		return
	}
	report, err := h.Mutations.ReadRGWSyncStatus(r.Context(), request.ClusterID, request.ZoneID, request.Name)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", report)
}
