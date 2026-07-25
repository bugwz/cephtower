package router

import "cephtower/backend/internal/api/v1/handler"

func settingRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/setting", h.ListSettings},
		{"PUT", "/setting", h.UpdateSettings},
		{"GET", "/setting/group", h.ListSettingGroups},
		{"GET", "/setting/group/{group}", h.GetSettingGroup},
		{"PUT", "/setting/group/{group}", h.UpdateSettingGroup},
		{"GET", "/setting/{name}", h.GetSetting},
		{"PUT", "/setting/{name}", h.UpdateSetting},
		{"DELETE", "/setting/{name}", h.ResetSetting},
	}
}
