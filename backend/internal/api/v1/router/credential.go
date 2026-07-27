package router

import "cephtower/backend/internal/api/v1/handler"

func credentialRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/credentials", h.ListCredentials},
		{"PUT", "/credential", h.PutCredential},
		{"DELETE", "/credential", h.DeleteCredential},
	}
}
