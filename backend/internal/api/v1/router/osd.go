package router

import "cephtower/backend/internal/api/v1/handler"

func osdRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/osd", h.ListOSDs},
		{"GET", "/osd/flag", h.OSDFlags},
		{"GET", "/osd/{id}", h.OSDDetails},
		{"GET", "/osd/{id}/device", h.ProxyPath},
		{"GET", "/osd/{id}/histogram", h.ProxyPath},
		{"PUT", "/osd/{id}/mark", h.ProxyPath},
		{"POST", "/osd/{id}/reweight", h.ProxyPath},
		{"POST", "/osd/{id}/scrub", h.ProxyPath},
		{"GET", "/osd/{id}/smart", h.ProxyPath},
	}
}
