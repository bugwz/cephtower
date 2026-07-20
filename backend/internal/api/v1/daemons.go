package v1

import (
	"net/http"

	"cephtower/backend/internal/integrations/ceph"
)

func registerDaemonRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/daemons", api.listDaemons)
	mux.HandleFunc("PUT /api/v1/ceph/daemons/{name}/action", api.applyDaemonAction)
}

func (api *API) listDaemons(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.ListDaemons(r.Context(), r.URL.Query().Get("types"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) applyDaemonAction(w http.ResponseWriter, r *http.Request) {
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
