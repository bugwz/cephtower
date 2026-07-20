package v1

import (
	"net/http"
)

func registerLogRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/logs", api.proxyCephGET("/api/logs/all"))
}
