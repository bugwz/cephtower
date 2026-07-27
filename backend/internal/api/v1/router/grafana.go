package router

import "cephtower/backend/internal/api/v1/handler"

func grafanaRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/grafana", h.GetGrafana},
	}
}
