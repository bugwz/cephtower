package v1

import (
	"net/http"

	"cephtower/backend/internal/integrations/ceph"
)

func registerHostRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/hosts", api.listHosts)
	mux.HandleFunc("POST /api/v1/ceph/hosts", api.createHost)
	mux.HandleFunc("GET /api/v1/ceph/hosts/{hostname}", api.hostDetails)
	mux.HandleFunc("PUT /api/v1/ceph/hosts/{hostname}", api.updateHost)
	mux.HandleFunc("DELETE /api/v1/ceph/hosts/{hostname}", api.deleteHost)
	mux.HandleFunc("GET /api/v1/ceph/hosts/{hostname}/daemons", api.hostDaemons)
	mux.HandleFunc("GET /api/v1/ceph/hosts/{hostname}/devices", api.hostDevices)
	mux.HandleFunc("GET /api/v1/ceph/hosts/{hostname}/inventory", api.hostInventory)
}

func (api *API) listHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := api.ceph.ListHosts(r.Context(), ceph.ListHostsOptions{
		Sources:                 r.URL.Query().Get("sources"),
		Facts:                   boolQuery(r.URL.Query(), "facts"),
		Offset:                  intQuery(r.URL.Query(), "offset"),
		Limit:                   intQuery(r.URL.Query(), "limit"),
		Search:                  r.URL.Query().Get("search"),
		Sort:                    r.URL.Query().Get("sort"),
		IncludeServiceInstances: boolQuery(r.URL.Query(), "include_service_instances"),
	})
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, hosts)
}

func (api *API) createHost(w http.ResponseWriter, r *http.Request) {
	var request ceph.HostRequest
	if !decodeRequestJSON(w, r, &request) {
		return
	}

	if err := api.ceph.CreateHost(r.Context(), request); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) hostDetails(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.HostDetails(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) updateHost(w http.ResponseWriter, r *http.Request) {
	var request ceph.UpdateHostRequest
	if !decodeRequestJSON(w, r, &request) {
		return
	}

	if err := api.ceph.UpdateHost(r.Context(), r.PathValue("hostname"), request); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) deleteHost(w http.ResponseWriter, r *http.Request) {
	if err := api.ceph.DeleteHost(r.Context(), r.PathValue("hostname")); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) hostDaemons(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.HostDaemons(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) hostDevices(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.HostDevices(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) hostInventory(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.HostInventory(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}
