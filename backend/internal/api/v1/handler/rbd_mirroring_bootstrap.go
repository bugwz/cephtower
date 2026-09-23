package handler

import (
	"net/http"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func (h *Handler) CreateRBDMirroringBootstrapToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Pool      string `json:"pool"`
		SiteName  string `json:"site_name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rbd_mirroring.bootstrap.create", "rbd_mirroring", request.Pool, "high", &request.ClusterID)
	if h.Mutations == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "RBD mirroring bootstrap is unavailable", false, nil)
		return
	}
	token, err := h.Mutations.CreateRBDMirrorBootstrapToken(r.Context(), request.ClusterID, request.Pool, request.SiteName)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, http.StatusOK, "success", map[string]string{"token": token})
}

func (h *Handler) ImportRBDMirroringBootstrapToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		Pool      string `json:"pool"`
		SiteName  string `json:"site_name"`
		Direction string `json:"direction"`
		Token     string `json:"token"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rbd_mirroring.bootstrap.import", "rbd_mirroring", request.Pool, "high", &request.ClusterID)
	if h.Mutations == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "RBD mirroring bootstrap is unavailable", false, nil)
		return
	}
	if err := h.Mutations.ImportRBDMirrorBootstrapToken(r.Context(), request.ClusterID, request.Pool, request.SiteName, request.Direction, request.Token); err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, http.StatusOK, "success", cephdomain.ActionResult{Details: map[string]any{"imported": true}})
}
