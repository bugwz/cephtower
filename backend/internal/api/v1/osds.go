package v1

import (
	"net/http"

	"cephtower/backend/internal/integrations/ceph"
)

func registerOSDRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/osds", api.listOSDs)
	mux.HandleFunc("GET /api/v1/ceph/osds/flags", api.osdFlags)
	mux.HandleFunc("GET /api/v1/ceph/osds/{id}", api.osdDetails)
	mux.HandleFunc("GET /api/v1/ceph/osds/{id}/devices", api.proxyCephGETPath("/api/osd/{id}/devices", "id"))
	mux.HandleFunc("GET /api/v1/ceph/osds/{id}/histogram", api.proxyCephGETPath("/api/osd/{id}/histogram", "id"))
	mux.HandleFunc("PUT /api/v1/ceph/osds/{id}/mark", api.proxyCephPath(http.MethodPut, "/api/osd/{id}/mark", "id"))
	mux.HandleFunc("POST /api/v1/ceph/osds/{id}/reweight", api.proxyCephPath(http.MethodPost, "/api/osd/{id}/reweight", "id"))
	mux.HandleFunc("POST /api/v1/ceph/osds/{id}/scrub", api.proxyCephPath(http.MethodPost, "/api/osd/{id}/scrub", "id"))
	mux.HandleFunc("GET /api/v1/ceph/osds/{id}/smart", api.proxyCephGETPath("/api/osd/{id}/smart", "id"))
}

func (api *API) listOSDs(w http.ResponseWriter, r *http.Request) {
	osds, err := api.ceph.ListOSDs(r.Context(), ceph.ListOSDsOptions{
		Offset: intQuery(r.URL.Query(), "offset"),
		Limit:  intQuery(r.URL.Query(), "limit"),
		Search: r.URL.Query().Get("search"),
		Sort:   r.URL.Query().Get("sort"),
	})
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, osds)
}

func (api *API) osdDetails(w http.ResponseWriter, r *http.Request) {
	payload, err := api.ceph.GetOSD(r.Context(), r.PathValue("id"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (api *API) osdFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := api.ceph.OSDFlags(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]string{"flags": flags})
}
