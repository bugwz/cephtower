package router

import "cephtower/backend/internal/api/v1/handler"

func managerRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/manager/telemetry/status", h.GetTelemetryStatus},
		{"GET", "/manager/telemetry/report", h.GetTelemetryReport},
		{"PATCH", "/manager/telemetry", h.UpdateTelemetry},
		{"PATCH", "/manager/telemetry/channel", h.UpdateTelemetryChannel},
		{"GET", "/managers", h.ListManagers},
		{"POST", "/manager/fail", h.FailManager},
	}
}
