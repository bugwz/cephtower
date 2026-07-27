package router

import "cephtower/backend/internal/api/v1/handler"

func healthRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/healthz", h.Health},
		{"GET", "/readyz", h.Ready},
	}
}
