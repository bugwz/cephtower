package router

import "cephtower/backend/internal/api/v1/handler"

func setupRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/setup/status", h.SetupStatus},
		{"POST", "/setup/initialize", h.InitializeSetup},
	}
}
