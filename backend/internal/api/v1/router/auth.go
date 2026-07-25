package router

import "cephtower/backend/internal/api/v1/handler"

func authRoutes(h *handler.Handler) []Route {
	return []Route{
		{"POST", "/auth/login", h.Login},
		{"POST", "/auth/password-reset/request", h.RequestPasswordReset},
		{"POST", "/auth/password-reset/confirm", h.ConfirmPasswordReset},
		{"GET", "/auth/me", h.Me},
		{"GET", "/user", h.ListUsers},
		{"POST", "/user", h.CreateUser},
		{"PATCH", "/user/{id}", h.UpdateUser},
	}
}
