package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	clusterservice "cephtower/backend/internal/service/cluster"
	service "cephtower/backend/internal/service/collector"
	"cephtower/backend/internal/store"
)

type CephClusterResponse struct {
	ID          uint                        `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	FSID        string                      `json:"fsid"`
	Enabled     bool                        `json:"enabled"`
	Dashboard   DashboardConnectionResponse `json:"dashboard"`
	Command     CommandConnectionResponse   `json:"command"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

type DashboardConnectionResponse struct {
	Enabled     bool   `json:"enabled"`
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	PasswordSet bool   `json:"password_set"`
	InsecureTLS bool   `json:"insecure_tls"`
}

type CommandConnectionResponse struct {
	Enabled           bool   `json:"enabled"`
	Bin               string `json:"bin"`
	Cluster           string `json:"cluster"`
	Conf              string `json:"conf"`
	MonitorHost       string `json:"monitor_host"`
	Name              string `json:"name"`
	Keyring           string `json:"keyring"`
	KeyringContentSet bool   `json:"keyring_content_set"`
	TimeoutSeconds    int    `json:"timeout_seconds"`
}

type CephClusterRequest struct {
	Name              string `json:"name"`
	MonitorHost       string `json:"monitor_host"`
	Keyring           string `json:"keyring"`
	DashboardUsername string `json:"dashboard_username"`
	DashboardPassword string `json:"dashboard_password"`
}

type CephClusterDetailResponse struct {
	Cluster   CephClusterResponse                `json:"cluster"`
	Discovery CephClusterDiscoveryDetailResponse `json:"discovery"`
}

type CephClusterDiscoveryDetailResponse struct {
	Hosts         []CephDiscoveredRecordResponse    `json:"hosts"`
	OSDs          []CephDiscoveredRecordResponse    `json:"osds"`
	OSDFlags      []CephClusterOSDFlagEntryResponse `json:"osd_flags"`
	Daemons       []CephDiscoveredRecordResponse    `json:"daemons"`
	Services      []CephDiscoveredRecordResponse    `json:"services"`
	Mons          []CephDiscoveredRecordResponse    `json:"mons"`
	Mgrs          []CephDiscoveredRecordResponse    `json:"mgrs"`
	MDSs          []CephDiscoveredRecordResponse    `json:"mdss"`
	MgrModules    []CephDiscoveredRecordResponse    `json:"mgr_modules"`
	Configuration []CephDiscoveredRecordResponse    `json:"configuration"`
}

type CephDiscoveredRecordResponse struct {
	Key          string    `json:"key"`
	Type         string    `json:"type,omitempty"`
	Hostname     string    `json:"hostname,omitempty"`
	Status       string    `json:"status,omitempty"`
	Payload      any       `json:"payload"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

type CephClusterOSDFlagEntryResponse struct {
	Name         string    `json:"name"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

type ClusterVersionResponse struct {
	Version string `json:"version"`
}

func (h *Handler) ListClusters(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	clusters, err := h.clusters.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := make([]CephClusterResponse, 0, len(clusters))
	for _, cluster := range clusters {
		response = append(response, toCephClusterResponse(cluster))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var req CephClusterRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.clusters.Create(r.Context(), clusterInput(req)); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, MessageResponse{Message: "集群连接已创建"})
}

func (h *Handler) GetCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	id, ok := clusterID(w, r)
	if !ok {
		return
	}
	detail, err := h.clusters.Get(r.Context(), id)
	if err != nil {
		writeClusterError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, CephClusterDetailResponse{
		Cluster:   toCephClusterResponse(detail.Cluster),
		Discovery: toDiscoveryResponse(detail.Discovery),
	})
}

func (h *Handler) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	id, ok := clusterID(w, r)
	if !ok {
		return
	}

	var req CephClusterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.clusters.Update(r.Context(), id, clusterInput(req)); err != nil {
		writeClusterError(w, err, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "集群连接已更新"})
}

func (h *Handler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	id, ok := clusterID(w, r)
	if !ok {
		return
	}

	if err := h.clusters.Delete(r.Context(), id); err != nil {
		writeClusterError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "集群连接已删除"})
}

func (h *Handler) ClusterSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.ceph.ClusterSummary(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) ClusterVersion(w http.ResponseWriter, r *http.Request) {
	version, err := h.ceph.Version(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ClusterVersionResponse{Version: version})
}

func (h *Handler) ClusterHealthMinimal(w http.ResponseWriter, r *http.Request) {
	health, err := h.ceph.HealthMinimal(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, health)
}

func clusterID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cluster id"})
		return 0, false
	}
	return uint(id), true
}

func clusterInput(req CephClusterRequest) clusterservice.Input {
	return clusterservice.Input{
		Name:              req.Name,
		MonitorHost:       req.MonitorHost,
		Keyring:           req.Keyring,
		DashboardUsername: req.DashboardUsername,
		DashboardPassword: req.DashboardPassword,
	}
}

func writeClusterError(w http.ResponseWriter, err error, fallbackStatus int) {
	if errors.Is(err, clusterservice.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, fallbackStatus, map[string]string{"error": err.Error()})
}

func toDiscoveryResponse(discovery clusterservice.Discovery) CephClusterDiscoveryDetailResponse {
	detail := CephClusterDiscoveryDetailResponse{}
	for _, host := range discovery.Hosts {
		detail.Hosts = append(detail.Hosts, CephDiscoveredRecordResponse{
			Key:          host.Hostname,
			Hostname:     host.Hostname,
			Status:       host.Status,
			Payload:      jsonPayload(host.Payload),
			DiscoveredAt: host.DiscoveredAt,
		})
	}

	for _, osd := range discovery.OSDs {
		detail.OSDs = append(detail.OSDs, CephDiscoveredRecordResponse{
			Key:          osd.OSDID,
			Hostname:     osd.Hostname,
			Status:       osd.Status,
			Payload:      jsonPayload(osd.Payload),
			DiscoveredAt: osd.DiscoveredAt,
		})
	}

	for _, flag := range discovery.OSDFlags {
		detail.OSDFlags = append(detail.OSDFlags, CephClusterOSDFlagEntryResponse{
			Name:         flag.Name,
			DiscoveredAt: flag.DiscoveredAt,
		})
	}

	for _, daemon := range discovery.Daemons {
		detail.Daemons = append(detail.Daemons, CephDiscoveredRecordResponse{
			Key:          daemon.Name,
			Type:         daemon.DaemonType,
			Hostname:     daemon.Hostname,
			Status:       daemon.Status,
			Payload:      jsonPayload(daemon.Payload),
			DiscoveredAt: daemon.DiscoveredAt,
		})
	}

	for _, service := range discovery.Services {
		detail.Services = append(detail.Services, CephDiscoveredRecordResponse{
			Key:          service.ServiceName,
			Type:         service.ServiceType,
			Payload:      jsonPayload(service.Payload),
			DiscoveredAt: service.DiscoveredAt,
		})
	}

	for _, mon := range discovery.Mons {
		detail.Mons = append(detail.Mons, CephDiscoveredRecordResponse{
			Key:          mon.Name,
			Type:         mon.Rank,
			Status:       mon.Status,
			Payload:      jsonPayload(mon.Payload),
			DiscoveredAt: mon.DiscoveredAt,
		})
	}

	for _, mgr := range discovery.Mgrs {
		detail.Mgrs = append(detail.Mgrs, CephDiscoveredRecordResponse{
			Key:          mgr.Name,
			Hostname:     mgr.Hostname,
			Status:       mgr.Status,
			Payload:      jsonPayload(mgr.Payload),
			DiscoveredAt: mgr.DiscoveredAt,
		})
	}

	for _, mds := range discovery.MDSs {
		detail.MDSs = append(detail.MDSs, CephDiscoveredRecordResponse{
			Key:          mds.Name,
			Type:         mds.Filesystem,
			Hostname:     mds.Hostname,
			Status:       mds.State,
			Payload:      jsonPayload(mds.Payload),
			DiscoveredAt: mds.DiscoveredAt,
		})
	}

	for _, module := range discovery.MgrModules {
		status := "disabled"
		if module.Enabled {
			status = "enabled"
		}
		detail.MgrModules = append(detail.MgrModules, CephDiscoveredRecordResponse{
			Key:          module.Name,
			Status:       status,
			Payload:      jsonPayload(module.Payload),
			DiscoveredAt: module.DiscoveredAt,
		})
	}

	for _, config := range discovery.Configuration {
		detail.Configuration = append(detail.Configuration, CephDiscoveredRecordResponse{
			Key:          strings.TrimSpace(config.Who + " " + config.Name),
			Type:         config.Level,
			Payload:      jsonPayload(config.Payload),
			DiscoveredAt: config.DiscoveredAt,
		})
	}

	return detail
}

func jsonPayload(payload string) any {
	var decoded any
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return payload
	}
	return decoded
}

func toCephClusterResponse(cluster store.CephCluster) CephClusterResponse {
	return CephClusterResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: "",
		FSID:        "",
		Enabled:     true,
		Dashboard: DashboardConnectionResponse{
			Enabled:     true,
			BaseURL:     "",
			Username:    cluster.DashboardUsername,
			Password:    cluster.DashboardPassword,
			PasswordSet: cluster.DashboardPassword != "",
			InsecureTLS: false,
		},
		Command: CommandConnectionResponse{
			Enabled:           true,
			Bin:               service.DefaultCephCommandBin,
			Cluster:           "",
			Conf:              "",
			MonitorHost:       cluster.MonitorHost,
			Name:              service.DefaultCephCommandName,
			Keyring:           cluster.Keyring,
			KeyringContentSet: cluster.Keyring != "",
			TimeoutSeconds:    service.DefaultCephCommandTimeoutSeconds,
		},
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

func (h *Handler) ClusterHealthFull(w http.ResponseWriter, r *http.Request) {
	health, err := h.ceph.HealthFull(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, health)
}
