package router

import "cephtower/backend/internal/api/v1/handler"

func daemonRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/daemon", h.ListDaemons},
		{"PUT", "/daemon/{name}/action", h.ApplyDaemonAction},
	}
}
