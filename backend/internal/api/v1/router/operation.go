package router

import "cephtower/backend/internal/api/v1/handler"

func operationRoutes(h *handler.Handler) []Route {
	return []Route{
		{"POST", "/operation/plan", h.CreatePlan},
		{"GET", "/operations", h.ListOperations},
		{"GET", "/operation", h.GetOperation},
		{"POST", "/operation/cancel", h.CancelOperation},
		{"POST", "/operation/retry", h.RetryOperation},
		{"GET", "/operation/events", h.OperationEvents},
		{"GET", "/operation/event/stream", h.StreamEvents},
	}
}
