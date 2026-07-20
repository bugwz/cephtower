package v1

import (
	"net/http"
)

func registerServiceRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/services", api.proxyCephGET("/api/service"))
	mux.HandleFunc("POST /api/v1/ceph/services", api.proxyCephPOST("/api/service"))
	mux.HandleFunc("GET /api/v1/ceph/services/known-types", api.proxyCephGET("/api/service/known_types"))
	mux.HandleFunc("GET /api/v1/ceph/services/{name}", api.proxyCephPath(http.MethodGet, "/api/service/{name}", "name"))
	mux.HandleFunc("PUT /api/v1/ceph/services/{name}", api.proxyCephPath(http.MethodPut, "/api/service/{name}", "name"))
	mux.HandleFunc("DELETE /api/v1/ceph/services/{name}", api.proxyCephPath(http.MethodDelete, "/api/service/{name}", "name"))
	mux.HandleFunc("GET /api/v1/ceph/services/{name}/daemons", api.proxyCephGETPath("/api/service/{name}/daemons", "name"))
}
