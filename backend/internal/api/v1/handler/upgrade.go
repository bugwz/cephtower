package handler

import "net/http"

func (h *Handler) ListUpgradeVersions(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "upgrade.versions", "upgrade", "versions", "", &request.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	if !h.ensureResourceCapability(w, r, request.ClusterID, "upgrade") {
		return
	}
	result, err := h.Inspection.UpgradeVersions(r.Context(), request.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}
