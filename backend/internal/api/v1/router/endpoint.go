package router

import "cephtower/backend/internal/api/v1/handler"

func endpointRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/endpoints", h.ListEndpoints},
		{"POST", "/endpoint", h.CreateEndpoint},
		{"PATCH", "/endpoint", h.UpdateEndpoint},
		{"DELETE", "/endpoint", h.DeleteEndpoint},
	}
}
