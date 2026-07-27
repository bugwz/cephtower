package router

import "cephtower/backend/internal/api/v1/handler"

func poolRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/pools", h.ListPools},
		{"POST", "/pool", h.CreatePool},
		{"GET", "/pool", h.GetPool},
		{"PATCH", "/pool", h.UpdatePool},
		{"DELETE", "/pool", h.DeletePool},
	}
}
