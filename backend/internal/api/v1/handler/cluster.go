package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
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
	ClusterID        uint64     `json:"cluster_id"`
	Name             string     `json:"name"`
	MonitorAddresses string     `json:"monitor_addresses"`
	ClientUsername   string     `json:"client_username"`
	FSID             string     `json:"fsid,omitempty"`
	CephVersion      string     `json:"ceph_version,omitempty"`
	Status           string     `json:"status"`
	Enabled          bool       `json:"enabled"`
	Generation       uint64     `json:"generation"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"`
	LastErrorCode    string     `json:"last_error_code,omitempty"`
	LastErrorMessage string     `json:"last_error_message,omitempty"`
	ObservedAt       *time.Time `json:"observed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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
	if r.URL.Query().Get("filter_options") == "1" {
		options, err := h.Clusters.FilterOptions(r.Context(), clusterFilterOptionFields(r))
		if err != nil {
			WriteError(w, r, 500, "store_error", err.Error(), false, nil)
			return
		}
		WriteSuccess(w, 200, "success", map[string]any{"filter_options": options})
		return
	}
	rows, err := h.Clusters.List(r.Context(), clusterFilters(r))
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

func clusterFilters(r *http.Request) store.ClusterFilter {
	return store.ClusterFilter{
		Names:           cleanClusterFilterValues(r.URL.Query()["filter.name"]),
		ClientUsernames: cleanClusterFilterValues(r.URL.Query()["filter.client_username"]),
	}
}

func clusterFilterOptionFields(r *http.Request) []string {
	requested := strings.Split(r.URL.Query().Get("fields"), ",")
	fields := make([]string, 0, len(requested))
	seen := map[string]bool{}
	for _, value := range requested {
		field := strings.TrimSpace(value)
		if seen[field] || (field != "name" && field != "client_username") {
			continue
		}
		seen[field] = true
		fields = append(fields, field)
	}
	return fields
}

func cleanClusterFilterValues(values []string) []string {
	cleaned := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func (h *Handler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	var request clusterCreateRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	cluster, result, err := h.Clusters.Create(r.Context(), clusterservice.CreateInput{Name: request.Name, MonitorAddresses: request.MonitorAddresses, ClientUsername: request.ClientUsername, ClientKey: request.ClientKey})
	if err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]any{"cluster": toClusterDTO(cluster), "result": result})
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
	result, err := h.Clusters.Update(r.Context(), id, clusterservice.UpdateInput{Name: request.Name, MonitorAddresses: request.MonitorAddresses, ClientUsername: request.ClientUsername, ClientKey: request.ClientKey})
	if err != nil {
		clusterError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", result)
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
	result, err := h.Clusters.Delete(r.Context(), id, request.DeleteCachedData)
	if err != nil {
		clusterError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", result)
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
	result, err := h.Clusters.Probe(r.Context(), id)
	if err != nil {
		clusterError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", result)
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
	dto := clusterDTO{
		ClusterID:        row.ID,
		Name:             row.Name,
		MonitorAddresses: row.MonitorAddresses,
		ClientUsername:   row.ClientUsername,
		Status:           row.Status,
		Enabled:          row.Enabled,
		Generation:       row.Generation,
		LastSeenAt:       row.LastSeenAt,
		ObservedAt:       row.ObservedAt,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	if row.FSID != nil {
		dto.FSID = *row.FSID
	}
	if row.CephVersion != nil {
		dto.CephVersion = cephdomain.NormalizeVersion(*row.CephVersion)
	}
	if row.LastErrorCode != nil {
		dto.LastErrorCode = *row.LastErrorCode
	}
	if row.LastErrorMessage != nil {
		dto.LastErrorMessage = *row.LastErrorMessage
	}
	applyClusterDiscovery(&dto, row.DiscoveredData)
	return dto
}

func applyClusterDiscovery(dto *clusterDTO, raw string) {
	var data map[string]any
	if json.Unmarshal([]byte(raw), &data) != nil {
		return
	}
	if value, ok := data["fsid"].(string); ok && value != "" {
		dto.FSID = value
	}
	version, _ := data["ceph_version"].(string)
	if version == "" {
		version, _ = data["version"].(string)
	}
	if version != "" {
		dto.CephVersion = cephdomain.NormalizeVersion(version)
	}
	if value, ok := data["status"].(string); ok && value != "" {
		dto.Status = value
	}
	if value, ok := data["error_code"].(string); ok {
		dto.LastErrorCode = value
	}
	if value, ok := data["error_message"].(string); ok {
		dto.LastErrorMessage = value
	}
}

func clusterError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, clusterservice.ErrNotFound) {
		WriteError(w, r, 404, "cluster_not_found", "cluster was not found", false, nil)
		return
	}
	var actionError *cephdomain.ActionError
	if errors.As(err, &actionError) {
		writeActionError(w, r, err)
		return
	}
	WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
}
