package router

import "cephtower/backend/internal/api/v1/handler"

func metricRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/metric/query", h.QueryMetric},
		{"GET", "/metric/range", h.QueryMetricRange},
	}
}
