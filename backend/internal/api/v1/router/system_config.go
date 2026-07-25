package router

import "cephtower/backend/internal/api/v1/handler"

func systemConfigRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/system/config/setting", h.ListSystemSettings},
		{"PUT", "/system/config/setting/{key}", h.UpdateSystemSetting},
		{"POST", "/system/config/default/reset", h.ResetSystemConfigDefaults},
		{"POST", "/system/config/data-fetch/{module}/run", h.RunDataFetchModuleNow},
		{"GET", "/system/config/data-fetch/run", h.ListDataFetchRuns},
	}
}
