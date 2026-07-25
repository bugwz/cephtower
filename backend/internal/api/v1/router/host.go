package router

import "cephtower/backend/internal/api/v1/handler"

func hostRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/host", h.ListHosts},
		{"POST", "/host", h.CreateHost},
		{"GET", "/host/{hostname}", h.HostDetails},
		{"PUT", "/host/{hostname}", h.UpdateHost},
		{"DELETE", "/host/{hostname}", h.DeleteHost},
		{"GET", "/host/{hostname}/daemon", h.HostDaemons},
		{"GET", "/host/{hostname}/device", h.HostDevices},
		{"GET", "/host/{hostname}/inventory", h.HostInventory},
	}
}
