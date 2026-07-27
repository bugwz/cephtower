package router

import "cephtower/backend/internal/api/v1/handler"

func erasureCodeProfileRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/erasure/code/profiles", h.ListErasureCodeProfiles},
		{"POST", "/erasure/code/profile", h.CreateErasureCodeProfile},
		{"GET", "/erasure/code/profile", h.GetErasureCodeProfile},
		{"DELETE", "/erasure/code/profile", h.DeleteErasureCodeProfile},
	}
}
