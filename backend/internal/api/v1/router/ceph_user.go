package router

import "cephtower/backend/internal/api/v1/handler"

func cephUserRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/ceph/users", h.ListCephUsers},
		{"POST", "/ceph/user", h.CreateCephUser},
		{"PATCH", "/ceph/user", h.UpdateCephUser},
		{"DELETE", "/ceph/user", h.DeleteCephUser},
		{"POST", "/ceph/users/import", h.ImportCephUsers},
		{"GET", "/ceph/users/export", h.ExportCephUsers},
	}
}
