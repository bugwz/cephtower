package router

import "cephtower/backend/internal/api/v1/handler"

func authRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/bootstrap", h.BootstrapStatus},
		{"POST", "/bootstrap/admin", h.BootstrapAdmin},
		{"POST", "/auth/login", h.Login},
		{"GET", "/user", h.ListUsers},
		{"POST", "/user", h.CreateUser},
	}
}
