package router

import "cephtower/backend/internal/api/v1/handler"

func rbacRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/role", h.ListRoles},
		{"POST", "/role", h.CreateRole},
		{"GET", "/role/bindings", h.ListRoleBindings},
		{"POST", "/role/binding", h.CreateRoleBinding},
		{"DELETE", "/role/binding", h.DeleteRoleBinding},
	}
}
