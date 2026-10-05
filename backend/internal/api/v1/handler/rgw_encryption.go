package handler

import "net/http"

func (h *Handler) GetRGWEncryptionConfiguration(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID      uint64 `json:"cluster_id"`
		Entity         string `json:"entity"`
		EncryptionType string `json:"encryption_type"`
		Provider       string `json:"provider"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rgw.encryption.read", "rgw_configuration", request.Entity, "", &request.ClusterID)
	w.Header().Set("Cache-Control", "no-store")
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RGWEncryptionConfiguration(r.Context(), request.ClusterID, request.Entity, request.EncryptionType, request.Provider)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}
