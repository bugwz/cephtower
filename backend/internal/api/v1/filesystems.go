package v1

import (
	"net/http"
)

func registerFilesystemRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/filesystems", api.proxyCephGET("/api/cephfs"))
	mux.HandleFunc("POST /api/v1/ceph/filesystems", api.proxyCephPOST("/api/cephfs"))
	mux.HandleFunc("GET /api/v1/ceph/filesystems/{id}", api.proxyCephGETPath("/api/cephfs/{id}", "id"))
	mux.HandleFunc("GET /api/v1/ceph/filesystems/{id}/clients", api.proxyCephGETPath("/api/cephfs/{id}/clients", "id"))
	mux.HandleFunc("GET /api/v1/ceph/filesystems/{id}/root", api.proxyCephGETPath("/api/cephfs/{id}/get_root_directory", "id"))
	mux.HandleFunc("GET /api/v1/ceph/filesystems/{id}/directory", api.proxyCephGETPath("/api/cephfs/{id}/ls_dir", "id"))
	mux.HandleFunc("GET /api/v1/ceph/filesystems/{id}/quota", api.proxyCephGETPath("/api/cephfs/{id}/quota", "id"))
	mux.HandleFunc("PUT /api/v1/ceph/filesystems/{id}/quota", api.proxyCephPath(http.MethodPut, "/api/cephfs/{id}/quota", "id"))
	mux.HandleFunc("GET /api/v1/ceph/filesystems/{id}/statfs", api.proxyCephGETPath("/api/cephfs/{id}/statfs", "id"))
}
