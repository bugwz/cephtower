package router

import "cephtower/backend/internal/api/v1/handler"

func monitorRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/monitors", h.ListMonitors},
		{"POST", "/monitor/action", h.RunMonitorAction},
	}
}
