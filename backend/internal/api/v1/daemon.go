package v1

import (
	"net/http"

	"cephtower/backend/internal/integrations/ceph"
)

func (api *API) ListDaemons(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.ListDaemons(r.Context(), r.URL.Query().Get("types"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) ApplyDaemonAction(w http.ResponseWriter, r *http.Request) {
	var request ceph.DaemonActionRequest
	if !decodeRequestJSON(w, r, &request) {
		return
	}

	if err := api.ceph.ApplyDaemonAction(r.Context(), r.PathValue("name"), request); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
