package router

import "cephtower/backend/internal/api/v1/handler"

func dashboardIdentityRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/dashboard-user", h.ListDashboardUsers},
		{"POST", "/dashboard-user", h.CreateDashboardUser},
		{"POST", "/dashboard-user/validate-password", h.ValidateDashboardUserPassword},
		{"GET", "/dashboard-user/{username}", h.GetDashboardUser},
		{"PUT", "/dashboard-user/{username}", h.UpdateDashboardUser},
		{"DELETE", "/dashboard-user/{username}", h.DeleteDashboardUser},
		{"POST", "/dashboard-user/{username}/change-password", h.ChangeDashboardUserPassword},
		{"GET", "/dashboard-role", h.ListDashboardRoles},
		{"POST", "/dashboard-role", h.CreateDashboardRole},
		{"GET", "/dashboard-role/{name}", h.GetDashboardRole},
		{"PUT", "/dashboard-role/{name}", h.UpdateDashboardRole},
		{"DELETE", "/dashboard-role/{name}", h.DeleteDashboardRole},
		{"POST", "/dashboard-role/{name}/clone", h.CloneDashboardRole},
	}
}
