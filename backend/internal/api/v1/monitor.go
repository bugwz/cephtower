package v1

import "net/http"

func registerMonitorRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/monitors", api.proxyCephGET("/api/monitor"))
}
