package router

import "cephtower/backend/internal/api/v1/handler"

func deviceRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/devices", h.ListDevices},
		{"POST", "/device/identify", h.IdentifyDevice},
		{"POST", "/device/zap", h.ZapDevice},
	}
}
