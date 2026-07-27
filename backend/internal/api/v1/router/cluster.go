package router

import "cephtower/backend/internal/api/v1/handler"

func clusterRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/clusters", h.ListClusters},
		{"POST", "/cluster", h.CreateCluster},
		{"GET", "/cluster", h.GetCluster},
		{"PATCH", "/cluster", h.UpdateCluster},
		{"DELETE", "/cluster", h.DeleteCluster},
		{"POST", "/cluster/probe", h.ProbeCluster},
		{"POST", "/cluster/refresh", h.RefreshCluster},
		{"GET", "/cluster/capabilities", h.Capabilities},
	}
}
