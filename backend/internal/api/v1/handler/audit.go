package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"cephtower/backend/internal/store"
)

type auditEventDTO struct {
	ID               uint64    `json:"audit_event_id"`
	OccurredAt       time.Time `json:"occurred_at"`
	EventType        string    `json:"event_type"`
	RequestID        string    `json:"request_id"`
	ActorUsername    string    `json:"actor_username"`
	ClusterID        *uint64   `json:"cluster_id"`
	ClusterName      *string   `json:"cluster_name"`
	Action           string    `json:"action"`
	ResourceKind     *string   `json:"resource_kind"`
	ResourceKey      *string   `json:"resource_key"`
	Risk             *string   `json:"risk"`
	Outcome          string    `json:"outcome"`
	HTTPStatus       *int      `json:"http_status"`
	ErrorCode        *string   `json:"error_code"`
	BeforeGeneration *uint64   `json:"before_generation"`
	AfterGeneration  *uint64   `json:"after_generation"`
	Parameters       any       `json:"parameters"`
	Details          any       `json:"details"`
	EventHash        string    `json:"event_hash"`
}

func (h *Handler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	actorUserID, ok := optionalUintQuery(w, r, "user_id")
	if !ok {
		return
	}
	rows, err := h.Database().ListAuditEventsFiltered(r.Context(), id, store.AuditFilter{ActorUsername: r.URL.Query().Get("username"), Action: r.URL.Query().Get("action"), ResourceKind: r.URL.Query().Get("resource_kind"), ResourceKey: r.URL.Query().Get("resource_key"), ActorUserID: actorUserID, Limit: limit})
	if err != nil {
		WriteError(w, r, 500, "store_error", err.Error(), false, nil)
		return
	}
	items := make([]auditEventDTO, 0, len(rows))
	for _, row := range rows {
		var parameters, details any
		if row.ParametersJSON != nil {
			_ = json.Unmarshal([]byte(*row.ParametersJSON), &parameters)
		}
		if row.DetailsJSON != nil {
			_ = json.Unmarshal([]byte(*row.DetailsJSON), &details)
		}
		items = append(items, auditEventDTO{ID: row.ID, OccurredAt: row.OccurredAt, EventType: row.EventType, RequestID: row.RequestID, ActorUsername: row.ActorUsername, ClusterID: row.ClusterID, ClusterName: row.ClusterName, Action: row.Action, ResourceKind: row.ResourceKind, ResourceKey: row.ResourceKey, Risk: row.Risk, Outcome: row.Outcome, HTTPStatus: row.HTTPStatus, ErrorCode: row.ErrorCode, BeforeGeneration: row.BeforeGeneration, AfterGeneration: row.AfterGeneration, Parameters: parameters, Details: details, EventHash: row.EventHash})
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
