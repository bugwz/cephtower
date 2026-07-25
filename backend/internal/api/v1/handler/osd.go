package handler

import (
	"net/http"

	ceph "cephtower/backend/internal/service/cephproxy"
)

func (h *Handler) ListOSDs(w http.ResponseWriter, r *http.Request) {
	osds, err := h.ceph.ListOSDs(r.Context(), ceph.ListOSDsOptions{
		Offset: intQuery(r.URL.Query(), "offset"),
		Limit:  intQuery(r.URL.Query(), "limit"),
		Search: r.URL.Query().Get("search"),
		Sort:   r.URL.Query().Get("sort"),
	})
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, osds)
}

func (h *Handler) OSDDetails(w http.ResponseWriter, r *http.Request) {
	payload, err := h.ceph.GetOSD(r.Context(), r.PathValue("id"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) OSDFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.ceph.OSDFlags(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, OSDFlagsResponse{Flags: flags})
}

type ListOSDsRequest = ceph.ListOSDsOptions
type OSDResponse = map[string]any
type OSDFlagsResponse struct {
	Flags []string `json:"flags"`
}
