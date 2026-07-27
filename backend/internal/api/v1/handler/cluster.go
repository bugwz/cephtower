package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/store"
)

type clusterCreateRequest struct {
	Name             string `json:"name"`
	MonitorAddresses string `json:"monitor_addresses"`
	ClientUsername   string `json:"client_username"`
	ClientKey        string `json:"client_key"`
}

type clusterUpdateRequest struct {
	ClusterID        uint64  `json:"cluster_id"`
	Name             *string `json:"name,omitempty"`
	MonitorAddresses *string `json:"monitor_addresses,omitempty"`
	ClientUsername   *string `json:"client_username,omitempty"`
	ClientKey        *string `json:"client_key,omitempty"`
}

type clusterDeleteRequest struct {
	ClusterID        uint64 `json:"cluster_id"`
	DeleteCachedData bool   `json:"delete_cached_data"`
}

type clusterIDRequest struct {
	ClusterID uint64 `json:"cluster_id"`
}

type clusterDTO struct {
	ClusterID        uint64    `json:"cluster_id"`
	Name             string    `json:"name"`
	MonitorAddresses string    `json:"monitor_addresses"`
	ClientUsername   string    `json:"client_username"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type capabilityDTO struct {
	Name       string    `json:"name"`
	Supported  bool      `json:"supported"`
	Reason     *string   `json:"reason"`
	Version    *string   `json:"version"`
	Details    any       `json:"details"`
	ObservedAt time.Time `json:"observed_at"`
}

func (h *Handler) ListClusters(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Clusters.List(r.Context())
	if err != nil {
		WriteError(w, r, 500, "store_error", err.Error(), false, nil)
		return
	}
	items := make([]clusterDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toClusterDTO(row))
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]string{"request_id": RequestID(r)}})
}

func (h *Handler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	var request clusterCreateRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	user, _ := CurrentUser(r)
	_, operation, err := h.Clusters.Create(r.Context(), clusterservice.CreateInput{Name: request.Name, MonitorAddresses: request.MonitorAddresses, ClientUsername: request.ClientUsername, ClientKey: request.ClientKey}, &user.ID, user.Username, RequestID(r), r.Header.Get("Idempotency-Key"))
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	w.Header().Set("Location", "/api/v1/operation")
	WriteSuccess(w, 202, "accepted", operationDTO(operation))
}

func (h *Handler) GetCluster(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	row, err := h.Clusters.Get(r.Context(), id)
	if err != nil {
		clusterError(w, r, err)
		return
	}
	WriteSuccess(w, 200, "success", toClusterDTO(row))
}

func (h *Handler) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	var request clusterUpdateRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	user, _ := CurrentUser(r)
	operation, err := h.Clusters.Update(r.Context(), id, clusterservice.UpdateInput{Name: request.Name, MonitorAddresses: request.MonitorAddresses, ClientUsername: request.ClientUsername, ClientKey: request.ClientKey}, &user.ID, user.Username, RequestID(r), r.Header.Get("Idempotency-Key"))
	if err != nil {
		clusterError(w, r, err)
		return
	}
	acceptedOperation(w, r, id, operation)
}

func (h *Handler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	var request clusterDeleteRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 || !request.DeleteCachedData {
		WriteError(w, r, 400, "invalid_request", "cluster_id and delete_cached_data=true are required", false, nil)
		return
	}
	user, _ := CurrentUser(r)
	operation, err := h.Clusters.Delete(r.Context(), id, request.DeleteCachedData, &user.ID, user.Username, RequestID(r), r.Header.Get("Idempotency-Key"))
	if err != nil {
		clusterError(w, r, err)
		return
	}
	acceptedOperation(w, r, id, operation)
}

func (h *Handler) ProbeCluster(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	user, _ := CurrentUser(r)
	operation, err := h.Clusters.Probe(r.Context(), id, &user.ID, user.Username, RequestID(r), r.Header.Get("Idempotency-Key"))
	if err != nil {
		clusterError(w, r, err)
		return
	}
	acceptedOperation(w, r, id, operation)
}

func (h *Handler) Capabilities(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	rows, err := h.Clusters.Capabilities(r.Context(), id)
	if err != nil {
		clusterError(w, r, err)
		return
	}
	items := make([]capabilityDTO, 0, len(rows))
	for _, row := range rows {
		var details any
		if row.DetailsJSON != nil {
			_ = json.Unmarshal([]byte(*row.DetailsJSON), &details)
		}
		items = append(items, capabilityDTO{Name: row.Name, Supported: row.Supported, Reason: row.Reason, Version: row.Version, Details: details, ObservedAt: row.ObservedAt})
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]string{"request_id": RequestID(r)}})
}

func toClusterDTO(row store.CephCluster) clusterDTO {
	return clusterDTO{ClusterID: row.ID, Name: row.Name, MonitorAddresses: row.MonitorAddresses, ClientUsername: row.ClientUsername, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func clusterError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, clusterservice.ErrNotFound) {
		WriteError(w, r, 404, "cluster_not_found", "cluster was not found", false, nil)
		return
	}
	WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
}
