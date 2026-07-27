package router

import "cephtower/backend/internal/api/v1/handler"

func serviceRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/services", h.ListServices},
		{"POST", "/service", h.CreateService},
		{"GET", "/service", h.GetService},
		{"PATCH", "/service", h.UpdateService},
		{"DELETE", "/service", h.DeleteService},
	}
}
