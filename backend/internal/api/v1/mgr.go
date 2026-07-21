package v1

import "net/http"

func registerMgrRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/mgr/modules", api.proxyCephGET("/api/mgr/module"))
	mux.HandleFunc("POST /api/v1/ceph/mgr/modules/{name}/enable", api.proxyCephPath(http.MethodPost, "/api/mgr/module/{name}/enable", "name"))
	mux.HandleFunc("POST /api/v1/ceph/mgr/modules/{name}/disable", api.proxyCephPath(http.MethodPost, "/api/mgr/module/{name}/disable", "name"))
}
