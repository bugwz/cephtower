package router

import "cephtower/backend/internal/api/v1/handler"

func rbdRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/rbd/images", h.ListRBDImages},
		{"POST", "/rbd/image", h.CreateRBDImage},
		{"GET", "/rbd/image", h.GetRBDImage},
		{"PATCH", "/rbd/image", h.UpdateRBDImage},
		{"DELETE", "/rbd/image", h.DeleteRBDImage},
		{"POST", "/rbd/image/action", h.RunRBDImageAction},
		{"GET", "/rbd/image/snapshots", h.ListRBDSnapshots},
		{"POST", "/rbd/image/snapshot", h.CreateRBDSnapshot},
		{"PATCH", "/rbd/image/snapshot", h.UpdateRBDSnapshot},
		{"DELETE", "/rbd/image/snapshot", h.DeleteRBDSnapshot},
		{"POST", "/rbd/image/snapshot/action", h.RunRBDSnapshotAction},
		{"GET", "/rbd/namespaces", h.ListRBDNamespaces},
		{"POST", "/rbd/namespace", h.CreateRBDNamespace},
		{"DELETE", "/rbd/namespace", h.DeleteRBDNamespace},
		{"GET", "/rbd/trash", h.ListRBDTrash},
		{"POST", "/rbd/trash/restore", h.RestoreRBDTrash},
		{"DELETE", "/rbd/trash", h.DeleteRBDTrash},
		{"POST", "/rbd/trash/purge", h.PurgeRBDTrash},
		{"GET", "/rbd/groups", h.ListRBDGroups},
		{"POST", "/rbd/group", h.CreateRBDGroup},
		{"GET", "/rbd/mirroring", h.GetRBDMirroring},
		{"PATCH", "/rbd/mirroring", h.UpdateRBDMirroring},
	}
}
