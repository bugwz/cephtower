package router

import "cephtower/backend/internal/api/v1/handler"

func prometheusRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/prometheus/settings", h.GetPrometheusSettings},
		{"PUT", "/prometheus/settings", h.UpdatePrometheusSettings},
		{"GET", "/prometheus", h.GetPrometheus},
		{"GET", "/prometheus/alertgroup", h.ListPrometheusAlertGroups},
		{"GET", "/prometheus/data", h.GetPrometheusData},
		{"GET", "/prometheus/notifications", h.ListPrometheusNotifications},
		{"GET", "/prometheus/query", h.QueryPrometheus},
		{"GET", "/prometheus/rules", h.ListPrometheusRules},
		{"GET", "/prometheus/silences", h.ListPrometheusSilences},
		{"POST", "/prometheus/silence", h.CreatePrometheusSilence},
		{"DELETE", "/prometheus/silence/{id}", h.DeletePrometheusSilence},
	}
}
