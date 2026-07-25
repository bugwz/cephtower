package handler

import "net/http"

// Dashboard-wide operational endpoints.
func (h *Handler) ListFeatureToggles(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/feature_toggles")
}

func (h *Handler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/feedback")
}

func (h *Handler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/feedback")
}

func (h *Handler) GetFeedbackAPIKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/feedback/api_key")
}

func (h *Handler) SetFeedbackAPIKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/feedback/api_key")
}

func (h *Handler) DeleteFeedbackAPIKey(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/feedback/api_key")
}

func (h *Handler) SetMOTD(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/motd")
}

func (h *Handler) ClearMOTD(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/motd/clear")
}

func (h *Handler) UpdateTelemetry(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/telemetry")
}

func (h *Handler) GetTelemetryReport(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/telemetry/report")
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/task")
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/logs/all")
}

type FeatureTogglesResponse = RawJSONResponse
type FeedbackRequest = RawJSONRequest
type FeedbackResponse = RawJSONResponse
type FeedbackAPIKeyRequest = RawJSONRequest
type MOTDRequest = RawJSONRequest
type TelemetryRequest = RawJSONRequest
type TelemetryResponse = RawJSONResponse
type TaskResponse = RawJSONResponse
type LogsResponse = RawJSONResponse
