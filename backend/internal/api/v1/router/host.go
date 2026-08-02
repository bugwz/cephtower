package router

import "cephtower/backend/internal/api/v1/handler"

func hostRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/hosts", h.ListHosts},
		{"POST", "/host", h.CreateHost},
		{"GET", "/host", h.GetHost},
		{"PATCH", "/host", h.UpdateHost},
		{"DELETE", "/host", h.DeleteHost},
		{"GET", "/host/ssh", h.GetHostSSH},
		{"PATCH", "/host/ssh", h.SaveHostSSH},
		{"GET", "/host/devices", h.ListHostDevices},
		{"POST", "/host/action", h.RunHostAction},
		{"POST", "/host/device/identify", h.IdentifyHostDevice},
		{"POST", "/host/device/zap", h.ZapDevice},
	}
}
