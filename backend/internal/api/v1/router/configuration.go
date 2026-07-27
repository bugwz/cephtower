package router

import "cephtower/backend/internal/api/v1/handler"

func configurationRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/configuration/options", h.ListConfigurationOptions},
		{"GET", "/configuration/values", h.ListConfigurationValues},
		{"PUT", "/configuration/value", h.SetConfigurationValue},
		{"DELETE", "/configuration/value", h.DeleteConfigurationValue},
	}
}
