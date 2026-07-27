package handler

import (
	"errors"
	"net/http"
	"strings"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func (h *Handler) ReadExternal(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, clusterID, ok := h.scopedBody(w, r)
		if !ok {
			return
		}
		if !h.ensureResourceCapability(w, r, clusterID, kind) {
			return
		}
		if h.External == nil {
			WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "external reader is unavailable", false, nil)
			return
		}
		key := readResourceKey(kind, body)
		if kind == "metric" {
			key = strings.TrimPrefix(r.URL.Path, "/api/v1/metric/")
		} else if kind == "nvmeof_namespace" && optionalStringBody(body, "nsid") != "" {
			key = optionalStringBody(body, "nqn", "subsystem_nqn") + "\x00" + optionalStringBody(body, "nsid")
		} else if kind == "nvmeof_listener" || kind == "nvmeof_host" || kind == "nvmeof_connection" {
			key = optionalStringBody(body, "nqn", "subsystem_nqn")
		}
		data, err := h.External.Read(r.Context(), clusterID, kind, key, r.URL.Query())
		if err != nil {
			var operationError *cephdomain.OperationError
			if errors.As(err, &operationError) {
				status := http.StatusBadGateway
				if operationError.Code == "invalid_request" || operationError.Code == "invalid_credential" || operationError.Code == "invalid_endpoint" {
					status = http.StatusBadRequest
				} else if operationError.Code == "endpoint_unavailable" || operationError.Code == "capability_unavailable" {
					status = http.StatusNotImplemented
				}
				WriteError(w, r, status, operationError.Code, operationError.Message, operationError.Retryable, nil)
				return
			}
			WriteError(w, r, http.StatusBadGateway, "external_unavailable", err.Error(), true, nil)
			return
		}
		WriteSuccess(w, http.StatusOK, "success", data)
	}
}
