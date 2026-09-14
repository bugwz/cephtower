package handler

import "net/http"

func (h *Handler) ListCephUsers(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("ceph_user", false)(w, r)
}
func (h *Handler) CreateCephUser(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("ceph_user", "ceph_user.create", "high")(w, r)
}
func (h *Handler) UpdateCephUser(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("ceph_user", "ceph_user.update", "high")(w, r)
}
func (h *Handler) DeleteCephUser(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("ceph_user", "ceph_user.delete", "high")(w, r)
}
func (h *Handler) ImportCephUsers(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("ceph_user", "ceph_user.import", "high")(w, r)
}

func (h *Handler) ExportCephUsers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ClusterID uint64   `json:"cluster_id"`
		Entities  []string `json:"entities"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "ceph_user.export", "ceph_user", "", "high", &request.ClusterID)
	if h.Mutations == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "Ceph user export is unavailable", false, nil)
		return
	}
	keyring, err := h.Mutations.ExportCephUsers(r.Context(), request.ClusterID, request.Entities)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteSuccess(w, http.StatusOK, "success", map[string]string{"keyring": keyring})
}
