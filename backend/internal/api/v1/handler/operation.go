package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type operationDTO struct {
	ID              uint64     `json:"operation_id"`
	ClusterID       uint64     `json:"cluster_id"`
	RequestID       string     `json:"request_id"`
	Action          string     `json:"action"`
	ResourceKind    string     `json:"resource_kind"`
	ResourceKey     string     `json:"resource_key"`
	Risk            string     `json:"risk"`
	Status          string     `json:"status"`
	ExpectedVersion *uint64    `json:"expected_version,omitempty"`
	Result          any        `json:"result,omitempty"`
	ErrorCode       *string    `json:"error_code,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	Retryable       bool       `json:"retryable"`
	Attempts        uint32     `json:"attempts"`
	MaxAttempts     uint32     `json:"max_attempts"`
	NextAttemptAt   *time.Time `json:"next_attempt_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (h *Handler) enqueueOperation(r *http.Request, request operationservice.EnqueueRequest) (store.CephOperation, error) {
	if h.Operations == nil {
		return store.CephOperation{}, errors.New("operation service is unavailable")
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(idempotencyKey) > 128 {
		return store.CephOperation{}, errInvalidIdempotencyKey
	}
	request.RequestID = RequestID(r)
	request.IdempotencyKey = idempotencyKey
	if user, ok := CurrentUser(r); ok && user.ID != 0 {
		request.ActorUserID = &user.ID
	}
	return h.Operations.Enqueue(r.Context(), request)
}

var errInvalidIdempotencyKey = errors.New("idempotency key must not exceed 128 characters")

func writeOperationEnqueueError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errInvalidIdempotencyKey):
		WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
	case errors.Is(err, operationservice.ErrIdempotencyConflict):
		WriteError(w, r, http.StatusConflict, "idempotency_conflict", err.Error(), false, nil)
	case err.Error() == "operation service is unavailable":
		WriteError(w, r, http.StatusServiceUnavailable, "not_ready", err.Error(), true, nil)
	default:
		WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), true, nil)
	}
}

func (h *Handler) GetOperation(w http.ResponseWriter, r *http.Request) {
	body, clusterID, ok := h.scopedBody(w, r)
	if !ok {
		return
	}
	operationID, ok := requiredUintBody(w, r, body, "operation_id")
	if !ok {
		return
	}
	annotateAudit(r, "operation.get", "operation", strconv.FormatUint(operationID, 10), "", &clusterID)
	row, err := h.Database().FindOperation(r.Context(), operationID)
	if errors.Is(err, store.ErrRecordNotFound) || (err == nil && row.ClusterID != clusterID) {
		WriteError(w, r, http.StatusNotFound, "operation_not_found", "operation was not found", false, nil)
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", toOperationDTO(row))
}

func (h *Handler) ListOperations(w http.ResponseWriter, r *http.Request) {
	_, clusterID, ok := h.scopedBody(w, r)
	if !ok {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && status != store.OperationQueued && status != store.OperationRunning &&
		status != store.OperationSucceeded && status != store.OperationFailed {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "status is invalid", false, nil)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil && r.URL.Query().Get("limit") != "" {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "limit must be an integer", false, nil)
		return
	}
	annotateAudit(r, "operation.list", "operation", "", "", &clusterID)
	rows, err := h.Database().ListOperations(r.Context(), store.OperationFilter{ClusterID: clusterID, Status: status, Limit: limit})
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", err.Error(), false, nil)
		return
	}
	items := make([]operationDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOperationDTO(row))
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]any{"items": items})
}

func toOperationDTO(row store.CephOperation) operationDTO {
	var result any
	if row.ResultJSON != nil {
		_ = json.Unmarshal([]byte(*row.ResultJSON), &result)
	}
	return operationDTO{
		ID: row.ID, ClusterID: row.ClusterID, RequestID: row.RequestID, Action: row.Action,
		ResourceKind: row.ResourceKind, ResourceKey: row.ResourceKey, Risk: row.Risk,
		Status: row.Status, ExpectedVersion: row.ExpectedVersion, Result: result,
		ErrorCode: row.ErrorCode, ErrorMessage: row.ErrorMessage, Retryable: row.Retryable,
		Attempts: row.Attempts, MaxAttempts: row.MaxAttempts, NextAttemptAt: row.NextAttemptAt,
		StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
