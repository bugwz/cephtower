package router

import "cephtower/backend/internal/api/v1/handler"

func upgradeRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/upgrade", h.GetUpgrade},
		{"POST", "/upgrade/check", h.CheckUpgrade},
		{"POST", "/upgrade/action", h.RunUpgradeAction},
	}
}
