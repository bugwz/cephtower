package router

import "cephtower/backend/internal/api/v1/handler"

func managerModuleRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/manager/modules", h.ListManagerModules},
		{"PATCH", "/manager/module", h.UpdateManagerModule},
	}
}
