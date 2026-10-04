package handler

import "net/http"

func (h *Handler) ReadRGWRealmToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var request struct {
		ClusterID uint64 `json:"cluster_id"`
		RealmID   string `json:"realm_id"`
		Name      string `json:"name"`
	}
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "rgw_realm.token.read", "rgw_realm", request.RealmID, "high", &request.ClusterID)
	if h.Mutations == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "Realm token reads are unavailable", false, nil)
		return
	}
	token, err := h.Mutations.ReadRGWRealmToken(r.Context(), request.ClusterID, request.RealmID, request.Name)
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]string{"token": token})
}
