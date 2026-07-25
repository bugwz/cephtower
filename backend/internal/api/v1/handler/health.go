package handler

import "net/http"

func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.setup == nil {
		writeJSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "not_ready"})
		return
	}
	initialized, _, err := h.setup.Status(r.Context())
	if err != nil || !initialized {
		writeJSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ready"})
}

type HealthResponse struct {
	Status string `json:"status"`
}
