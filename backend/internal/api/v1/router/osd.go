package router

import "cephtower/backend/internal/api/v1/handler"

func osdRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/osds", h.ListOSDs},
		{"GET", "/osd", h.GetOSD},
		{"GET", "/osd/inspection", h.GetOSDInspection},
		{"GET", "/osd/flag", h.GetOSDFlag},
		{"PATCH", "/osd/flag", h.UpdateOSDFlag},
		{"PATCH", "/osd/flag/individual", h.UpdateOSDIndividualFlag},
		{"POST", "/osd/action", h.RunOSDAction},
		{"PUT", "/osd/device/class", h.UpdateOSDDeviceClass},
		{"POST", "/osd/removal/check", h.CheckOSDRemoval},
		{"DELETE", "/osd", h.DeleteOSD},
		{"GET", "/osd/removals", h.ListOSDRemovals},
		{"POST", "/osd/deployment/preview", h.PreviewOSDDeployment},
		{"POST", "/osd/deployment", h.CreateOSDDeployment},
	}
}
