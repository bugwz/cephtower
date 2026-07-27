package router

import "cephtower/backend/internal/api/v1/handler"

func alertRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/alert/alerts", h.ListAlerts},
		{"GET", "/alert/rules", h.ListAlertRules},
		{"GET", "/alert/silences", h.ListSilences},
		{"POST", "/alert/silence", h.CreateSilence},
		{"DELETE", "/alert/silence", h.DeleteSilence},
	}
}
