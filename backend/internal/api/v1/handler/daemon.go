package handler

import (
	"net/http"

	ceph "cephtower/backend/internal/service/cephproxy"
)

func (h *Handler) ListDaemons(w http.ResponseWriter, r *http.Request) {
	payload, err := h.ceph.ListDaemons(r.Context(), r.URL.Query().Get("types"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) ApplyDaemonAction(w http.ResponseWriter, r *http.Request) {
	var request ceph.DaemonActionRequest
	if !decodeRequestJSON(w, r, &request) {
		return
	}

	if err := h.ceph.ApplyDaemonAction(r.Context(), r.PathValue("name"), request); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type ListDaemonsRequest struct {
	Types string `json:"types"`
}

type DaemonActionRequest = ceph.DaemonActionRequest
type DaemonResponse = map[string]any
