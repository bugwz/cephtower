package v1

import (
	"net/http"
)

func registerConfigurationRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/configuration", api.proxyCephGET("/api/cluster_conf"))
	mux.HandleFunc("POST /api/v1/ceph/configuration", api.proxyCephPOST("/api/cluster_conf"))
	mux.HandleFunc("PUT /api/v1/ceph/configuration", api.proxyCephPUT("/api/cluster_conf"))
	mux.HandleFunc("GET /api/v1/ceph/configuration/filter", api.proxyCephGET("/api/cluster_conf/filter"))
	mux.HandleFunc("GET /api/v1/ceph/configuration/{name}", api.proxyCephGETPath("/api/cluster_conf/{name}", "name"))
	mux.HandleFunc("DELETE /api/v1/ceph/configuration/{name}", api.proxyCephPath(http.MethodDelete, "/api/cluster_conf/{name}", "name"))
}
