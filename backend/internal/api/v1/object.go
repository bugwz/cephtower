package v1

import (
	"net/http"
)

func registerObjectRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/object/gateways", api.proxyCephGET("/api/rgw/daemon"))
	mux.HandleFunc("GET /api/v1/ceph/object/gateways/{id}", api.proxyCephGETPath("/api/rgw/daemon/{id}", "id"))
	mux.HandleFunc("GET /api/v1/ceph/object/users", api.proxyCephGET("/api/rgw/user"))
	mux.HandleFunc("POST /api/v1/ceph/object/users", api.proxyCephPOST("/api/rgw/user"))
	mux.HandleFunc("GET /api/v1/ceph/object/users/{uid}", api.proxyCephGETPath("/api/rgw/user/{uid}", "uid"))
	mux.HandleFunc("PUT /api/v1/ceph/object/users/{uid}", api.proxyCephPath(http.MethodPut, "/api/rgw/user/{uid}", "uid"))
	mux.HandleFunc("DELETE /api/v1/ceph/object/users/{uid}", api.proxyCephPath(http.MethodDelete, "/api/rgw/user/{uid}", "uid"))
	mux.HandleFunc("GET /api/v1/ceph/object/buckets", api.proxyCephGET("/api/rgw/bucket"))
	mux.HandleFunc("POST /api/v1/ceph/object/buckets", api.proxyCephPOST("/api/rgw/bucket"))
	mux.HandleFunc("GET /api/v1/ceph/object/buckets/{bucket}", api.proxyCephGETPath("/api/rgw/bucket/{bucket}", "bucket"))
	mux.HandleFunc("PUT /api/v1/ceph/object/buckets/{bucket}", api.proxyCephPath(http.MethodPut, "/api/rgw/bucket/{bucket}", "bucket"))
	mux.HandleFunc("DELETE /api/v1/ceph/object/buckets/{bucket}", api.proxyCephPath(http.MethodDelete, "/api/rgw/bucket/{bucket}", "bucket"))
	mux.HandleFunc("GET /api/v1/ceph/object/accounts", api.proxyCephGET("/api/rgw/accounts"))
	mux.HandleFunc("POST /api/v1/ceph/object/accounts", api.proxyCephPOST("/api/rgw/accounts"))
	mux.HandleFunc("GET /api/v1/ceph/object/accounts/{id}", api.proxyCephGETPath("/api/rgw/accounts/{id}", "id"))
	mux.HandleFunc("PUT /api/v1/ceph/object/accounts/{id}", api.proxyCephPath(http.MethodPut, "/api/rgw/accounts/{id}", "id"))
	mux.HandleFunc("DELETE /api/v1/ceph/object/accounts/{id}", api.proxyCephPath(http.MethodDelete, "/api/rgw/accounts/{id}", "id"))
}
