package router

import "cephtower/backend/internal/api/v1/handler"

func configurationRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/dashboard/configuration", h.ListConfiguration},
		{"POST", "/dashboard/configuration", h.CreateConfiguration},
		{"PUT", "/dashboard/configuration", h.UpdateConfiguration},
		{"GET", "/dashboard/configuration/filter", h.ListConfigurationFilter},
		{"GET", "/dashboard/configuration/{name}", h.GetConfiguration},
		{"DELETE", "/dashboard/configuration/{name}", h.DeleteConfiguration},
		{"GET", "/dashboard/mgr/module", h.ListMgrModules},
		{"GET", "/dashboard/mgr/module/{name}", h.GetMgrModule},
		{"PUT", "/dashboard/mgr/module/{name}", h.UpdateMgrModule},
		{"POST", "/dashboard/mgr/module/{name}/enable", h.EnableMgrModule},
		{"POST", "/dashboard/mgr/module/{name}/disable", h.DisableMgrModule},
		{"GET", "/dashboard/mgr/module/{name}/options", h.GetMgrModuleOptions},
	}
}
