package router

import "cephtower/backend/internal/api/v1/handler"

func logsRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/logs", h.ListLogs},
		{"GET", "/logs/stream", h.StreamEvents},
	}
}
