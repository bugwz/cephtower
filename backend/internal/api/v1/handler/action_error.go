package handler

import (
	"errors"
	"net/http"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func writeActionError(w http.ResponseWriter, r *http.Request, err error) {
	var actionError *cephdomain.ActionError
	if errors.As(err, &actionError) {
		status := http.StatusBadGateway
		switch actionError.Code {
		case "invalid_request", "invalid_credential", "invalid_endpoint":
			status = http.StatusBadRequest
		case "endpoint_unavailable", "capability_unavailable":
			status = http.StatusNotImplemented
		case "resource_conflict":
			status = http.StatusConflict
		}
		WriteError(w, r, status, actionError.Code, actionError.Message, actionError.Retryable, actionError.Details)
		return
	}
	WriteError(w, r, http.StatusBadGateway, "action_failed", err.Error(), true, nil)
}
