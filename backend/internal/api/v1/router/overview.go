package router

import "cephtower/backend/internal/api/v1/handler"

func overviewRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/overview", h.GetOverview},
		{"GET", "/health", h.ListHealthChecks},
		{"POST", "/health/mute", h.MuteHealthCheck},
		{"DELETE", "/health/mute", h.UnmuteHealthCheck},
	}
}
