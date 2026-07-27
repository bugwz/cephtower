package router

import "cephtower/backend/internal/api/v1/handler"

func smbRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/smb/clusters", h.ListSMBClusters},
		{"POST", "/smb/cluster", h.CreateSMBCluster},
		{"GET", "/smb/cluster", h.GetSMBCluster},
		{"PATCH", "/smb/cluster", h.UpdateSMBCluster},
		{"DELETE", "/smb/cluster", h.DeleteSMBCluster},
		{"GET", "/smb/shares", h.ListSMBShares},
		{"POST", "/smb/share", h.CreateSMBShare},
		{"GET", "/smb/share", h.GetSMBShare},
		{"PATCH", "/smb/share", h.UpdateSMBShare},
		{"DELETE", "/smb/share", h.DeleteSMBShare},
	}
}
