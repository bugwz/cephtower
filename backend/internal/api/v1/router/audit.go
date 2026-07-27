package router

import "cephtower/backend/internal/api/v1/handler"

func auditRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/audit/events", h.ListAuditEvents},
	}
}
