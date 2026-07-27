package router

import "cephtower/backend/internal/api/v1/handler"

func deviceRoutes(h *handler.Handler) []Route {
	return []Route{
		{"POST", "/device/zap", h.ZapDevice},
	}
}
