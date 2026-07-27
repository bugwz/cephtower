package handler

import "net/http"

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, http.StatusOK, "success", map[string]string{"status": "ok"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.Database == nil || h.Database() == nil {
		WriteError(w, r, http.StatusServiceUnavailable, "not_ready", "database is unavailable", true, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]string{"status": "ready"})
}
