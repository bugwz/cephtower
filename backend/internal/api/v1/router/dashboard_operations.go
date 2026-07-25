package router

import "cephtower/backend/internal/api/v1/handler"

func dashboardOperationsRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/feature-toggles", h.ListFeatureToggles},
		{"GET", "/feedback", h.GetFeedback},
		{"POST", "/feedback", h.SubmitFeedback},
		{"GET", "/feedback/api-key", h.GetFeedbackAPIKey},
		{"POST", "/feedback/api-key", h.SetFeedbackAPIKey},
		{"DELETE", "/feedback/api-key", h.DeleteFeedbackAPIKey},
		{"POST", "/motd", h.SetMOTD},
		{"DELETE", "/motd", h.ClearMOTD},
		{"PUT", "/telemetry", h.UpdateTelemetry},
		{"GET", "/telemetry/report", h.GetTelemetryReport},
		{"GET", "/task", h.ListTasks},
		{"GET", "/logs", h.ListLogs},
	}
}
