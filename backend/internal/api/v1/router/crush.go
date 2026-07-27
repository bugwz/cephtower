package router

import "cephtower/backend/internal/api/v1/handler"

func crushRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/crush/rules", h.ListCrushRules},
		{"POST", "/crush/rule", h.CreateCrushRule},
		{"GET", "/crush/rule", h.GetCrushRule},
		{"PATCH", "/crush/rule", h.UpdateCrushRule},
		{"DELETE", "/crush/rule", h.DeleteCrushRule},
	}
}
