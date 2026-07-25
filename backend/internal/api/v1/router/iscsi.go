package router

import "cephtower/backend/internal/api/v1/handler"

func iscsiRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/iscsi/settings", h.GetISCSISettings},
		{"PUT", "/iscsi/settings", h.UpdateISCSISettings},
		{"GET", "/iscsi/discoveryauth", h.GetISCSIDiscoveryAuth},
		{"PUT", "/iscsi/discoveryauth", h.UpdateISCSIDiscoveryAuth},
		{"GET", "/iscsi/target", h.ListISCSITargets},
		{"POST", "/iscsi/target", h.CreateISCSITarget},
		{"GET", "/iscsi/target/{iqn}", h.GetISCSITarget},
		{"PUT", "/iscsi/target/{iqn}", h.UpdateISCSITarget},
		{"DELETE", "/iscsi/target/{iqn}", h.DeleteISCSITarget},
	}
}
