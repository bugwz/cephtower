package router

import "cephtower/backend/internal/api/v1/handler"

func grafanaRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/grafana/settings", h.GetGrafanaSettings},
		{"PUT", "/grafana/settings", h.UpdateGrafanaSettings},
		{"GET", "/grafana/url", h.GetGrafanaURL},
		{"GET", "/grafana/validation/{params}", h.ValidateGrafana},
		{"POST", "/grafana/dashboards", h.SyncGrafanaDashboards},
	}
}
