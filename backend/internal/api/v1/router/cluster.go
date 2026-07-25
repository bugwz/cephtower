package router

import "cephtower/backend/internal/api/v1/handler"

func clusterRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/cluster", h.ListClusters},
		{"POST", "/cluster", h.CreateCluster},
		{"GET", "/cluster/{id}", h.GetCluster},
		{"PUT", "/cluster/{id}", h.UpdateCluster},
		{"DELETE", "/cluster/{id}", h.DeleteCluster},
		{"GET", "/cluster/summary", h.ClusterSummary},
		{"GET", "/cluster/version", h.ClusterVersion},
		{"GET", "/cluster/health", h.ClusterHealthMinimal},
		{"GET", "/cluster/health/full", h.ClusterHealthFull},
	}
}
