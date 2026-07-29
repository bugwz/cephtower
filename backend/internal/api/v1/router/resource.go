package router

import "cephtower/backend/internal/api/v1/handler"

func resourceRoutes(h *handler.Handler) []Route {
	return []Route{
		{"POST", "/resource/refresh", h.RefreshResource},
	}
}
