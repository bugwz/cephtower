package handler

import (
	"net/http"

	ceph "cephtower/backend/internal/service/cephproxy"
)

func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.ceph.ListHosts(r.Context(), ceph.ListHostsOptions{
		Sources:                 r.URL.Query().Get("sources"),
		Facts:                   boolQuery(r.URL.Query(), "facts"),
		Offset:                  intQuery(r.URL.Query(), "offset"),
		Limit:                   intQuery(r.URL.Query(), "limit"),
		Search:                  r.URL.Query().Get("search"),
		Sort:                    r.URL.Query().Get("sort"),
		IncludeServiceInstances: boolQuery(r.URL.Query(), "include_service_instances"),
	})
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, hosts)
}

func (h *Handler) CreateHost(w http.ResponseWriter, r *http.Request) {
	var request ceph.HostRequest
	if !decodeRequestJSON(w, r, &request) {
		return
	}

	if err := h.ceph.CreateHost(r.Context(), request); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HostDetails(w http.ResponseWriter, r *http.Request) {
	payload, err := h.ceph.HostDetails(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) UpdateHost(w http.ResponseWriter, r *http.Request) {
	var request ceph.UpdateHostRequest
	if !decodeRequestJSON(w, r, &request) {
		return
	}

	if err := h.ceph.UpdateHost(r.Context(), r.PathValue("hostname"), request); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	if err := h.ceph.DeleteHost(r.Context(), r.PathValue("hostname")); err != nil {
		writeCephError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HostDaemons(w http.ResponseWriter, r *http.Request) {
	payload, err := h.ceph.HostDaemons(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) HostDevices(w http.ResponseWriter, r *http.Request) {
	payload, err := h.ceph.HostDevices(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) HostInventory(w http.ResponseWriter, r *http.Request) {
	payload, err := h.ceph.HostInventory(r.Context(), r.PathValue("hostname"))
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

type ListHostsRequest = ceph.ListHostsOptions
type HostRequest = ceph.HostRequest
type UpdateHostRequest = ceph.UpdateHostRequest
type HostResponse = ceph.Host
type HostDetailResponse = map[string]any
type HostDaemonsResponse = []map[string]any
type HostDevicesResponse = []map[string]any
type HostInventoryResponse = map[string]any
