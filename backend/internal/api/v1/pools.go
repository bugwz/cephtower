package v1

import (
	"net/http"
)

func registerPoolRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/pools", api.proxyCephGET("/api/pool"))
	mux.HandleFunc("POST /api/v1/ceph/pools", api.proxyCephPOST("/api/pool"))
	mux.HandleFunc("GET /api/v1/ceph/pools/{name}", api.proxyCephPath(http.MethodGet, "/api/pool/{name}", "name"))
	mux.HandleFunc("PUT /api/v1/ceph/pools/{name}", api.proxyCephPath(http.MethodPut, "/api/pool/{name}", "name"))
	mux.HandleFunc("DELETE /api/v1/ceph/pools/{name}", api.proxyCephPath(http.MethodDelete, "/api/pool/{name}", "name"))
	mux.HandleFunc("GET /api/v1/ceph/pools/{name}/configuration", api.proxyCephGETPath("/api/pool/{name}/configuration", "name"))
}
