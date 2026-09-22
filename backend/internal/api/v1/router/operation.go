package router

import "cephtower/backend/internal/api/v1/handler"

func operationRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/operation", h.GetOperation},
		{"GET", "/operations", h.ListOperations},
	}
}
