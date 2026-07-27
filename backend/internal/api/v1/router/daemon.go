package router

import "cephtower/backend/internal/api/v1/handler"

func daemonRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/daemons", h.ListDaemons},
		{"GET", "/daemon", h.GetDaemon},
		{"POST", "/daemon/action", h.RunDaemonAction},
	}
}
