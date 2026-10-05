package handler

import "net/http"

func (h *Handler) GetRGWRealmBucketUsage(w http.ResponseWriter, r *http.Request) {
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
	annotateAudit(r, "rgw.realm.bucket.usage", "rgw_bucket", "", "", &input.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RGWRealmBucketUsage(r.Context(), input.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", result)
}

func (h *Handler) GetRGWBucketUsage(w http.ResponseWriter, r *http.Request) {
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
	annotateAudit(r, "rgw.bucket.usage", "rgw_bucket", "", "", &input.ClusterID)
	if h.Inspection == nil {
		WriteError(w, r, 501, "capability_unavailable", "cluster inspection is unavailable", false, nil)
		return
	}
	result, err := h.Inspection.RGWBucketUsage(r.Context(), input.ClusterID)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	result["scope"] = "current_rgw_configuration"
	WriteSuccess(w, 200, "success", result)
}
