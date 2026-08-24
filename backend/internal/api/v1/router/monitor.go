package router

import "cephtower/backend/internal/api/v1/handler"

func monitorRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/monitors", h.ListMonitors},
		{"GET", "/monitor/status", h.GetMonitorStatus},
		{"GET", "/monitor/perf/counters", h.ListMonitorPerfCounters},
		{"POST", "/monitor/action", h.RunMonitorAction},
	}
}
