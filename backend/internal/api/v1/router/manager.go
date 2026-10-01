package router

import "cephtower/backend/internal/api/v1/handler"

func managerRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/manager/telemetry/status", h.GetTelemetryStatus},
		{"GET", "/managers", h.ListManagers},
		{"POST", "/manager/fail", h.FailManager},
	}
}
