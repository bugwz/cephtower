package v1

import (
	"net/http"
)

func registerBlockRoutes(mux *http.ServeMux, api *API) {
	mux.HandleFunc("GET /api/v1/ceph/block/images", api.proxyCephGET("/api/block/image"))
	mux.HandleFunc("POST /api/v1/ceph/block/images", api.proxyCephPOST("/api/block/image"))
	mux.HandleFunc("GET /api/v1/ceph/block/images/default-features", api.proxyCephGET("/api/block/image/default_features"))
	mux.HandleFunc("GET /api/v1/ceph/block/images/clone-format-version", api.proxyCephGET("/api/block/image/clone_format_version"))
	mux.HandleFunc("GET /api/v1/ceph/block/images/trash", api.proxyCephGET("/api/block/image/trash"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/trash/purge", api.proxyCephPOST("/api/block/image/trash/purge"))
	mux.HandleFunc("DELETE /api/v1/ceph/block/images/trash/{image}", api.proxyCephPath(http.MethodDelete, "/api/block/image/trash/{image}", "image"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/trash/{image}/restore", api.proxyCephPath(http.MethodPost, "/api/block/image/trash/{image}/restore", "image"))
	mux.HandleFunc("GET /api/v1/ceph/block/images/{image}", api.proxyCephPath(http.MethodGet, "/api/block/image/{image}", "image"))
	mux.HandleFunc("PUT /api/v1/ceph/block/images/{image}", api.proxyCephPath(http.MethodPut, "/api/block/image/{image}", "image"))
	mux.HandleFunc("DELETE /api/v1/ceph/block/images/{image}", api.proxyCephPath(http.MethodDelete, "/api/block/image/{image}", "image"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/{image}/copy", api.proxyCephPath(http.MethodPost, "/api/block/image/{image}/copy", "image"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/{image}/flatten", api.proxyCephPath(http.MethodPost, "/api/block/image/{image}/flatten", "image"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/{image}/snapshots", api.proxyCephPath(http.MethodPost, "/api/block/image/{image}/snap", "image"))
	mux.HandleFunc("DELETE /api/v1/ceph/block/images/{image}/snapshots/{snapshot}", api.proxyCephPath2(http.MethodDelete, "/api/block/image/{image}/snap/{snapshot}", "image", "snapshot"))
	mux.HandleFunc("PUT /api/v1/ceph/block/images/{image}/snapshots/{snapshot}", api.proxyCephPath2(http.MethodPut, "/api/block/image/{image}/snap/{snapshot}", "image", "snapshot"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/{image}/snapshots/{snapshot}/clone", api.proxyCephPath2(http.MethodPost, "/api/block/image/{image}/snap/{snapshot}/clone", "image", "snapshot"))
	mux.HandleFunc("POST /api/v1/ceph/block/images/{image}/snapshots/{snapshot}/rollback", api.proxyCephPath2(http.MethodPost, "/api/block/image/{image}/snap/{snapshot}/rollback", "image", "snapshot"))
	mux.HandleFunc("GET /api/v1/ceph/block/mirroring/summary", api.proxyCephGET("/api/block/mirroring/summary"))
}
