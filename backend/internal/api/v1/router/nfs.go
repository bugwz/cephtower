package router

import "cephtower/backend/internal/api/v1/handler"

func nfsRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/nfs/settings", h.GetNFSSettings},
		{"PUT", "/nfs/settings", h.UpdateNFSSettings},
		{"GET", "/nfs/cluster", h.ListNFSClusters},
		{"GET", "/nfs/export", h.ListNFSExports},
		{"POST", "/nfs/export", h.CreateNFSExport},
		{"DELETE", "/nfs/export/{cluster_id}/{id}", h.DeleteNFSExport},
	}
}
