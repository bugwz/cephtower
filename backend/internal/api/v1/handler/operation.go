package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type operationEventDTO struct {
	Sequence  uint64    `json:"sequence"`
	EventType string    `json:"event_type"`
	Stage     string    `json:"stage"`
	Progress  *int      `json:"progress"`
	Message   string    `json:"message"`
	Data      any       `json:"data"`
	ErrorCode *string   `json:"error_code"`
	CreatedAt time.Time `json:"created_at"`
}

type planDTO struct {
	ID                 string    `json:"plan_id"`
	ClusterID          uint64    `json:"cluster_id"`
	Action             string    `json:"action"`
	ResourceKind       string    `json:"resource_kind"`
	ResourceKey        string    `json:"resource_key"`
	ResourceGeneration uint64    `json:"resource_generation"`
	Risk               string    `json:"risk"`
	Status             string    `json:"status"`
	Blockers           []string  `json:"blockers"`
	Warnings           []string  `json:"warnings"`
	ExpiresAt          time.Time `json:"expires_at"`
	CreatedAt          time.Time `json:"created_at"`
}

func operationDTO(row store.CephOperation) cephdomain.OperationView {
	view := cephdomain.OperationView{ID: row.ID, Action: row.Action, ClusterID: row.ClusterID, ResourceType: row.ResourceKind, ResourceKey: row.ResourceKey, Status: row.Status, Stage: row.Stage, Progress: row.Progress, Risk: row.Risk, CreatedAt: row.CreatedAt, StartedAt: row.StartedAt, CompletedAt: row.CompletedAt}
	if row.ResultJSON != nil {
		_ = json.Unmarshal([]byte(*row.ResultJSON), &view.Result)
	}
	if row.ErrorCode != nil {
		view.Error = map[string]any{"error_code": *row.ErrorCode, "message": row.ErrorMessage, "retryable": row.Retryable}
	}
	return view
}

func acceptedOperation(w http.ResponseWriter, r *http.Request, clusterID uint64, row store.CephOperation) {
	w.Header().Set("Location", "/api/v1/operation")
	WriteSuccess(w, 202, "accepted", operationDTO(row))
}

func (h *Handler) ListOperations(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	actorUserID, ok := optionalUintQuery(w, r, "user_id")
	if !ok {
		return
	}
	rows, err := h.Operations.ListFiltered(r.Context(), id, store.OperationFilter{Status: r.URL.Query().Get("status"), Action: r.URL.Query().Get("kind"), ResourceKind: r.URL.Query().Get("resource_kind"), ResourceKey: r.URL.Query().Get("resource_key"), ActorUserID: actorUserID, Limit: limit})
	if err != nil {
		WriteError(w, r, 500, "store_error", err.Error(), false, nil)
		return
	}
	items := make([]cephdomain.OperationView, 0, len(rows))
	for _, row := range rows {
		items = append(items, operationDTO(row))
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]string{"request_id": RequestID(r)}})
}

func optionalUintQuery(w http.ResponseWriter, r *http.Request, name string) (*uint64, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return nil, true
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", name+" must be a positive integer", false, nil)
		return nil, false
	}
	return &value, true
}

func (h *Handler) GetOperation(w http.ResponseWriter, r *http.Request) {
	request, ok := operationCommandRequest(w, r)
	if !ok {
		return
	}
	row, ok := h.scopedOperationByID(w, r, request.ClusterID, request.OperationID)
	if !ok {
		return
	}
	WriteSuccess(w, 200, "success", operationDTO(row))
}

func (h *Handler) OperationEvents(w http.ResponseWriter, r *http.Request) {
	request, ok := operationCommandRequest(w, r)
	if !ok {
		return
	}
	if _, ok := h.scopedOperationByID(w, r, request.ClusterID, request.OperationID); !ok {
		return
	}
	rows, err := h.Operations.Events(r.Context(), request.OperationID)
	if err != nil {
		WriteError(w, r, 500, "store_error", err.Error(), false, nil)
		return
	}
	items := make([]operationEventDTO, 0, len(rows))
	for _, row := range rows {
		var data any
		if row.DataJSON != nil {
			_ = json.Unmarshal([]byte(*row.DataJSON), &data)
		}
		items = append(items, operationEventDTO{Sequence: row.Sequence, EventType: row.EventType, Stage: row.Stage, Progress: row.Progress, Message: row.Message, Data: data, ErrorCode: row.ErrorCode, CreatedAt: row.CreatedAt})
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]string{"request_id": RequestID(r)}})
}

func (h *Handler) CancelOperation(w http.ResponseWriter, r *http.Request) {
	request, ok := operationCommandRequest(w, r)
	if !ok {
		return
	}
	old, ok := h.scopedOperationByID(w, r, request.ClusterID, request.OperationID)
	if !ok {
		return
	}
	if err := h.Operations.Cancel(r.Context(), request.OperationID); err != nil {
		WriteError(w, r, 409, "operation_conflict", err.Error(), false, nil)
		return
	}
	row, _ := h.Operations.Get(r.Context(), old.ID)
	acceptedOperation(w, r, request.ClusterID, row)
}

func (h *Handler) RetryOperation(w http.ResponseWriter, r *http.Request) {
	request, ok := operationCommandRequest(w, r)
	if !ok {
		return
	}
	old, ok := h.scopedOperationByID(w, r, request.ClusterID, request.OperationID)
	if !ok {
		return
	}
	if old.Status != "failed" || !old.Retryable {
		WriteError(w, r, 409, "operation_conflict", "operation is not retryable", false, nil)
		return
	}
	var parameters any
	_ = json.Unmarshal([]byte(old.RequestJSON), &parameters)
	user, _ := CurrentUser(r)
	retryID := old.ID
	row, err := h.Operations.Enqueue(r.Context(), operationservice.Request{ClusterID: old.ClusterID, ClusterName: old.ClusterName, ActorUserID: &user.ID, ActorUsername: user.Username, RequestID: RequestID(r), Action: old.Action, ResourceKind: old.ResourceKind, ResourceKey: old.ResourceKey, ResourceGeneration: old.ResourceGeneration, Risk: cephdomain.Risk(old.Risk), RetryOfID: &retryID, Parameters: parameters})
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	acceptedOperation(w, r, request.ClusterID, row)
}

type operationCommandBody struct {
	ClusterID   uint64 `json:"cluster_id"`
	OperationID string `json:"operation_id"`
}

func operationCommandRequest(w http.ResponseWriter, r *http.Request) (operationCommandBody, bool) {
	var request operationCommandBody
	if !DecodeStrict(w, r, &request) {
		return operationCommandBody{}, false
	}
	request.OperationID = strings.TrimSpace(request.OperationID)
	if request.ClusterID == 0 || request.OperationID == "" {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id and operation_id are required", false, nil)
		return operationCommandBody{}, false
	}
	return request, true
}

func (h *Handler) scopedOperationByID(w http.ResponseWriter, r *http.Request, clusterID uint64, operationID string) (store.CephOperation, bool) {
	row, err := h.Operations.Get(r.Context(), operationID)
	if err != nil || row.ClusterID == nil || *row.ClusterID != clusterID {
		WriteError(w, r, 404, "operation_not_found", "operation was not found", false, nil)
		return store.CephOperation{}, false
	}
	return row, true
}

type planRequest struct {
	ClusterID          uint64          `json:"cluster_id"`
	Action             string          `json:"action"`
	ResourceKind       string          `json:"resource_kind"`
	ResourceKey        string          `json:"resource_key"`
	ResourceGeneration uint64          `json:"resource_generation"`
	Risk               cephdomain.Risk `json:"risk"`
	Parameters         map[string]any  `json:"parameters"`
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var request planRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	if request.Parameters == nil {
		request.Parameters = map[string]any{}
	}
	request.Parameters["cluster_id"] = float64(id)
	if err := ValidateMutationRequest(request.Action, request.Parameters); err != nil {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
		return
	}
	user, _ := CurrentUser(r)
	plan, err := h.Operations.CreatePlan(r.Context(), operationservice.PlanRequest{ClusterID: id, ActorUserID: &user.ID, ActorUsername: user.Username, RequestID: RequestID(r), Action: request.Action, ResourceKind: request.ResourceKind, ResourceKey: request.ResourceKey, ResourceGeneration: request.ResourceGeneration, Risk: request.Risk, Parameters: request.Parameters})
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	var blockers, warnings []string
	_ = json.Unmarshal([]byte(plan.BlockersJSON), &blockers)
	_ = json.Unmarshal([]byte(plan.WarningsJSON), &warnings)
	WriteSuccess(w, 200, "success", planDTO{ID: plan.ID, ClusterID: plan.ClusterID, Action: plan.Action, ResourceKind: plan.ResourceKind, ResourceKey: plan.ResourceKey, ResourceGeneration: plan.ResourceGeneration, Risk: plan.Risk, Status: plan.Status, Blockers: blockers, Warnings: warnings, ExpiresAt: plan.ExpiresAt, CreatedAt: plan.CreatedAt})
}

func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	clusterID := request.ClusterID
	if clusterID == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, r, 500, "stream_unavailable", "streaming is unavailable", false, nil)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	cursor := uint64(0)
	if raw := r.Header.Get("Last-Event-ID"); raw != "" {
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			WriteError(w, r, http.StatusBadRequest, "invalid_request", "Last-Event-ID must be an unsigned integer", false, nil)
			return
		}
		cursor = value
	}
	poll := time.NewTicker(time.Second)
	heartbeat := time.NewTicker(10 * time.Second)
	defer poll.Stop()
	defer heartbeat.Stop()
	emit := func(eventType string, id uint64, data any) bool {
		response := APIResponse{Code: 0, Message: "success", Data: data}
		encoded, _ := json.Marshal(response)
		if id > 0 {
			if _, err := fmt.Fprintf(w, "id: %d\n", id); err != nil {
				return false
			}
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, encoded); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}
	emitPending := func() bool {
		if !strings.HasSuffix(r.URL.Path, "/operation/event/stream") || h.Database == nil || h.Database() == nil {
			return true
		}
		for {
			rows, err := h.Database().ListClusterOperationEventsAfter(r.Context(), clusterID, cursor, 100)
			if err != nil {
				return false
			}
			for _, row := range rows {
				var data any
				if row.DataJSON != nil {
					_ = json.Unmarshal([]byte(*row.DataJSON), &data)
				}
				item := operationEventDTO{Sequence: row.Sequence, EventType: row.EventType, Stage: row.Stage, Progress: row.Progress, Message: row.Message, Data: data, ErrorCode: row.ErrorCode, CreatedAt: row.CreatedAt}
				if !emit("operation", row.ID, map[string]any{"type": "operation_event", "operation_id": row.OperationID, "event": item}) {
					return false
				}
				cursor = row.ID
			}
			if len(rows) < 100 {
				return true
			}
		}
	}
	if !emitPending() {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case <-poll.C:
			if !emitPending() {
				return
			}
		case <-heartbeat.C:
			if !emit("heartbeat", 0, map[string]any{"type": "heartbeat", "observed_at": time.Now().UTC()}) {
				return
			}
		}
	}
}

func riskFromString(risk string) cephdomain.Risk {
	switch risk {
	case "high":
		return cephdomain.RiskHigh
	case "medium":
		return cephdomain.RiskMedium
	default:
		return cephdomain.RiskLow
	}
}

func requestResourceKey(r *http.Request) string {
	return strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/"), "/")
}
