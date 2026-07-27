package router

import "cephtower/backend/internal/api/v1/handler"

func iscsiRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/iscsi/gateway", h.GetISCSIGateway},
		{"GET", "/iscsi/targets", h.ListISCSITargets},
		{"POST", "/iscsi/target", h.CreateISCSITarget},
		{"GET", "/iscsi/target", h.GetISCSITarget},
		{"PATCH", "/iscsi/target", h.UpdateISCSITarget},
		{"DELETE", "/iscsi/target", h.DeleteISCSITarget},
	}
}
