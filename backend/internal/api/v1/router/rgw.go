package router

import "cephtower/backend/internal/api/v1/handler"

func rgwRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/rgw/settings", h.GetRGWSettings},
		{"PUT", "/rgw/settings", h.UpdateRGWSettings},
		{"GET", "/rgw/status", h.GetRGWStatus},
		{"POST", "/rgw/validate", h.ValidateRGWConfig},
	}
}
