package router

import "cephtower/backend/internal/api/v1/handler"

func nfsRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/nfs/clusters", h.ListNFSClusters},
		{"POST", "/nfs/cluster", h.CreateNFSCluster},
		{"GET", "/nfs/cluster", h.GetNFSCluster},
		{"DELETE", "/nfs/cluster", h.DeleteNFSCluster},
		{"GET", "/nfs/exports", h.ListNFSExports},
		{"POST", "/nfs/export", h.CreateNFSExport},
		{"GET", "/nfs/export", h.GetNFSExport},
		{"PATCH", "/nfs/export", h.UpdateNFSExport},
		{"DELETE", "/nfs/export", h.DeleteNFSExport},
	}
}
