package handler

import (
	"errors"
	"net/http"
	"time"

	endpointservice "cephtower/backend/internal/service/endpoint"
	"cephtower/backend/internal/store"
)

type credentialRequest struct {
	ClusterID  uint64         `json:"cluster_id"`
	Kind       string         `json:"kind"`
	Credential map[string]any `json:"credential"`
}

type credentialDeleteRequest struct {
	ClusterID uint64 `json:"cluster_id"`
	Kind      string `json:"kind"`
}

type credentialDTO struct {
	Kind        string    `json:"kind"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type endpointRequest struct {
	ClusterID      uint64  `json:"cluster_id"`
	EndpointID     uint64  `json:"endpoint_id,omitempty"`
	Kind           string  `json:"kind"`
	Name           string  `json:"name"`
	URL            string  `json:"url"`
	TLSMode        string  `json:"tls_mode"`
	CACredentialID *uint64 `json:"ca_credential_id"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	Enabled        *bool   `json:"enabled"`
}

type endpointDTO struct {
	ID             uint64    `json:"endpoint_id"`
	Kind           string    `json:"kind"`
	Name           string    `json:"name"`
	URL            string    `json:"url"`
	TLSMode        string    `json:"tls_mode"`
	CACredentialID *uint64   `json:"ca_credential_id"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (h *Handler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	rows, err := h.Endpoints.ListCredentials(r.Context(), id)
	if err != nil {
		WriteError(w, r, 500, "store_error", "could not list credentials", false, nil)
		return
	}
	items := make([]credentialDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, credentialDTO{Kind: row.Kind, Fingerprint: row.Fingerprint, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]any{"request_id": RequestID(r)}})
}

func (h *Handler) PutCredential(w http.ResponseWriter, r *http.Request) {
	var request credentialRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if request.ClusterID == 0 || request.Kind == "" {
		WriteError(w, r, 400, "invalid_request", "cluster_id and kind are required", false, nil)
		return
	}
	row, err := h.Endpoints.PutCredential(r.Context(), request.ClusterID, endpointservice.CredentialInput{Kind: request.Kind, Value: request.Credential})
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 200, "success", credentialDTO{Kind: row.Kind, Fingerprint: row.Fingerprint, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
}

func (h *Handler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	var request credentialDeleteRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if request.ClusterID == 0 || request.Kind == "" {
		WriteError(w, r, 400, "invalid_request", "cluster_id and kind are required", false, nil)
		return
	}
	err := h.Endpoints.DeleteCredential(r.Context(), request.ClusterID, request.Kind)
	if errors.Is(err, store.ErrRecordNotFound) {
		WriteError(w, r, 404, "credential_not_found", "credential was not found", false, nil)
		return
	}
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 200, "success", nil)
}

func (h *Handler) ListEndpoints(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	id := request.ClusterID
	if id == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	rows, err := h.Endpoints.ListEndpoints(r.Context(), id)
	if err != nil {
		WriteError(w, r, 500, "store_error", "could not list endpoints", false, nil)
		return
	}
	items := make([]endpointDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEndpointDTO(row))
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]any{"request_id": RequestID(r)}})
}

func (h *Handler) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var request endpointRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if request.ClusterID == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	row, err := h.Endpoints.CreateEndpoint(r.Context(), request.ClusterID, endpointInput(request))
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 201, "success", toEndpointDTO(row))
}

func (h *Handler) UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	var request endpointRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if request.ClusterID == 0 || request.EndpointID == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id and endpoint_id are required", false, nil)
		return
	}
	row, err := h.Endpoints.UpdateEndpoint(r.Context(), request.ClusterID, request.EndpointID, endpointInput(request))
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 200, "success", toEndpointDTO(row))
}

func (h *Handler) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var request endpointRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if request.ClusterID == 0 || request.EndpointID == 0 {
		WriteError(w, r, 400, "invalid_request", "cluster_id and endpoint_id are required", false, nil)
		return
	}
	err := h.Endpoints.DeleteEndpoint(r.Context(), request.ClusterID, request.EndpointID)
	if errors.Is(err, store.ErrRecordNotFound) {
		WriteError(w, r, 404, "endpoint_not_found", "endpoint was not found", false, nil)
		return
	}
	if err != nil {
		WriteError(w, r, 500, "store_error", "could not delete endpoint", false, nil)
		return
	}
	WriteSuccess(w, 200, "success", nil)
}

func endpointInput(value endpointRequest) endpointservice.EndpointInput {
	return endpointservice.EndpointInput{Kind: value.Kind, Name: value.Name, URL: value.URL, TLSMode: value.TLSMode, CACredentialID: value.CACredentialID, TimeoutSeconds: value.TimeoutSeconds, Enabled: value.Enabled}
}

func toEndpointDTO(row store.CephClusterEndpoint) endpointDTO {
	return endpointDTO{ID: row.ID, Kind: row.Kind, Name: row.Name, URL: row.URL, TLSMode: row.TLSMode, CACredentialID: row.CACredentialID, Enabled: row.Enabled, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
