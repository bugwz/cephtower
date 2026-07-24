package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/service/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type cephClusterResponse struct {
	ID          uint                        `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	FSID        string                      `json:"fsid"`
	Enabled     bool                        `json:"enabled"`
	Dashboard   dashboardConnectionResponse `json:"dashboard"`
	Command     commandConnectionResponse   `json:"command"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

type dashboardConnectionResponse struct {
	Enabled     bool   `json:"enabled"`
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	PasswordSet bool   `json:"password_set"`
	InsecureTLS bool   `json:"insecure_tls"`
}

type commandConnectionResponse struct {
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

type cephClusterRequest struct {
	Name              string `json:"name"`
	MonitorHost       string `json:"monitor_host"`
	Keyring           string `json:"keyring"`
	DashboardUsername string `json:"dashboard_username"`
	DashboardPassword string `json:"dashboard_password"`
}

type clusterActionResponse struct {
	Message string `json:"message"`
}

type cephClusterDetailResponse struct {
	Cluster   cephClusterResponse        `json:"cluster"`
	Discovery cephClusterDiscoveryDetail `json:"discovery"`
}

type cephClusterDiscoveryDetail struct {
	Hosts         []cephDiscoveredRecord    `json:"hosts"`
	OSDs          []cephDiscoveredRecord    `json:"osds"`
	OSDFlags      []cephClusterOSDFlagEntry `json:"osd_flags"`
	Daemons       []cephDiscoveredRecord    `json:"daemons"`
	Services      []cephDiscoveredRecord    `json:"services"`
	Mons          []cephDiscoveredRecord    `json:"mons"`
	Mgrs          []cephDiscoveredRecord    `json:"mgrs"`
	MDSs          []cephDiscoveredRecord    `json:"mdss"`
	MgrModules    []cephDiscoveredRecord    `json:"mgr_modules"`
	Configuration []cephDiscoveredRecord    `json:"configuration"`
}

type cephDiscoveredRecord struct {
	Key          string    `json:"key"`
	Type         string    `json:"type,omitempty"`
	Hostname     string    `json:"hostname,omitempty"`
	Status       string    `json:"status,omitempty"`
	Payload      any       `json:"payload"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

type cephClusterOSDFlagEntry struct {
	Name         string    `json:"name"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

func (api *API) ListClusters(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var clusters []store.CephCluster
	if err := api.database().Order("id asc").Find(&clusters).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := make([]cephClusterResponse, 0, len(clusters))
	for _, cluster := range clusters {
		response = append(response, toCephClusterResponse(cluster))
	}
	writeJSON(w, http.StatusOK, response)
}

func (api *API) CreateCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var req cephClusterRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	cluster, err := buildCephCluster(req, nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	err = api.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cluster).Error; err != nil {
			return err
		}
		if err := ceph.EnsureDefaultSystemSettings(r.Context(), tx); err != nil {
			return err
		}
		return api.discoverCephCluster(r.Context(), tx, &cluster)
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, clusterActionResponse{Message: "集群连接已创建"})
}

func (api *API) GetCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cluster, ok := api.clusterByID(w, r)
	if !ok {
		return
	}

	discovery, err := api.clusterDiscovery(r.Context(), cluster.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cephClusterDetailResponse{
		Cluster:   toCephClusterResponse(cluster),
		Discovery: discovery,
	})
}

func (api *API) GetClusterKeyring(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cluster, ok := api.clusterByID(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"keyring": cluster.Keyring})
}

func (api *API) GetClusterDashboardPassword(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cluster, ok := api.clusterByID(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"dashboard_password": cluster.DashboardPassword})
}

func (api *API) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	current, ok := api.clusterByID(w, r)
	if !ok {
		return
	}

	var req cephClusterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	cluster, err := buildCephCluster(req, &current)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	cluster.ID = current.ID
	cluster.CreatedAt = current.CreatedAt

	err = api.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&cluster).Error; err != nil {
			return err
		}
		if err := ceph.EnsureDefaultSystemSettings(r.Context(), tx); err != nil {
			return err
		}
		return api.discoverCephCluster(r.Context(), tx, &cluster)
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, clusterActionResponse{Message: "集群连接已更新"})
}

func (api *API) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cluster, ok := api.clusterByID(w, r)
	if !ok {
		return
	}

	err := api.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := ceph.DeleteCephClusterResources(r.Context(), tx, cluster.ID); err != nil {
			return err
		}
		return tx.Delete(&cluster).Error
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if api.clusterRuntimeCleaner != nil {
		if err := api.clusterRuntimeCleaner(r.Context(), cluster.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, clusterActionResponse{Message: "集群连接已删除"})
}

func (api *API) ClusterSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := api.ceph.ClusterSummary(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (api *API) ClusterVersion(w http.ResponseWriter, r *http.Request) {
	version, err := api.ceph.Version(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"version": version})
}

func (api *API) ClusterHealthMinimal(w http.ResponseWriter, r *http.Request) {
	health, err := api.ceph.HealthMinimal(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, health)
}

func (api *API) clusterByID(w http.ResponseWriter, r *http.Request) (store.CephCluster, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cluster id"})
		return store.CephCluster{}, false
	}

	var cluster store.CephCluster
	err = api.database().First(&cluster, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "cluster not found"})
		return store.CephCluster{}, false
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return store.CephCluster{}, false
	}
	return cluster, true
}

func buildCephCluster(req cephClusterRequest, current *store.CephCluster) (store.CephCluster, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return store.CephCluster{}, fmt.Errorf("name is required")
	}

	cluster := store.CephCluster{
		Name:              name,
		MonitorHost:       strings.TrimSpace(req.MonitorHost),
		DashboardUsername: strings.TrimSpace(req.DashboardUsername),
	}

	if current != nil {
		cluster.ID = current.ID
		cluster.CreatedAt = current.CreatedAt
		cluster.DashboardPassword = current.DashboardPassword
		cluster.Keyring = current.Keyring
		cluster.MonitorHost = current.MonitorHost
	}
	if strings.TrimSpace(req.MonitorHost) != "" {
		cluster.MonitorHost = strings.TrimSpace(req.MonitorHost)
	}
	if req.DashboardPassword != "" {
		cluster.DashboardPassword = req.DashboardPassword
	}
	if req.Keyring != "" {
		cluster.Keyring = req.Keyring
	}

	if cluster.DashboardUsername == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard username is required")
	}
	if strings.TrimSpace(cluster.MonitorHost) == "" {
		return store.CephCluster{}, fmt.Errorf("monitor host is required")
	}
	if cluster.DashboardPassword == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard password is required")
	}
	if strings.TrimSpace(cluster.Keyring) == "" {
		return store.CephCluster{}, fmt.Errorf("keyring is required")
	}

	return cluster, nil
}

func (api *API) discoverCephCluster(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
	if api.clusterDiscoverer == nil {
		return nil
	}
	return api.clusterDiscoverer(ctx, db, cluster)
}

func (api *API) clusterDiscovery(ctx context.Context, clusterID uint) (cephClusterDiscoveryDetail, error) {
	detail := cephClusterDiscoveryDetail{}

	var hosts []store.CephClusterHost
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("hostname asc").Find(&hosts).Error; err != nil {
		return detail, err
	}
	for _, host := range hosts {
		detail.Hosts = append(detail.Hosts, cephDiscoveredRecord{
			Key:          host.Hostname,
			Hostname:     host.Hostname,
			Status:       host.Status,
			Payload:      jsonPayload(host.Payload),
			DiscoveredAt: host.DiscoveredAt,
		})
	}

	var osds []store.CephClusterOSD
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("osd_id asc").Find(&osds).Error; err != nil {
		return detail, err
	}
	for _, osd := range osds {
		detail.OSDs = append(detail.OSDs, cephDiscoveredRecord{
			Key:          osd.OSDID,
			Hostname:     osd.Hostname,
			Status:       osd.Status,
			Payload:      jsonPayload(osd.Payload),
			DiscoveredAt: osd.DiscoveredAt,
		})
	}

	var flags []store.CephClusterOSDFlag
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&flags).Error; err != nil {
		return detail, err
	}
	for _, flag := range flags {
		detail.OSDFlags = append(detail.OSDFlags, cephClusterOSDFlagEntry{
			Name:         flag.Name,
			DiscoveredAt: flag.DiscoveredAt,
		})
	}

	var daemons []store.CephClusterDaemon
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&daemons).Error; err != nil {
		return detail, err
	}
	for _, daemon := range daemons {
		detail.Daemons = append(detail.Daemons, cephDiscoveredRecord{
			Key:          daemon.Name,
			Type:         daemon.DaemonType,
			Hostname:     daemon.Hostname,
			Status:       daemon.Status,
			Payload:      jsonPayload(daemon.Payload),
			DiscoveredAt: daemon.DiscoveredAt,
		})
	}

	var services []store.CephClusterService
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("service_name asc").Find(&services).Error; err != nil {
		return detail, err
	}
	for _, service := range services {
		detail.Services = append(detail.Services, cephDiscoveredRecord{
			Key:          service.ServiceName,
			Type:         service.ServiceType,
			Payload:      jsonPayload(service.Payload),
			DiscoveredAt: service.DiscoveredAt,
		})
	}

	var mons []store.CephClusterMon
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&mons).Error; err != nil {
		return detail, err
	}
	for _, mon := range mons {
		detail.Mons = append(detail.Mons, cephDiscoveredRecord{
			Key:          mon.Name,
			Type:         mon.Rank,
			Status:       mon.Status,
			Payload:      jsonPayload(mon.Payload),
			DiscoveredAt: mon.DiscoveredAt,
		})
	}

	var mgrs []store.CephClusterMgr
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&mgrs).Error; err != nil {
		return detail, err
	}
	for _, mgr := range mgrs {
		detail.Mgrs = append(detail.Mgrs, cephDiscoveredRecord{
			Key:          mgr.Name,
			Hostname:     mgr.Hostname,
			Status:       mgr.Status,
			Payload:      jsonPayload(mgr.Payload),
			DiscoveredAt: mgr.DiscoveredAt,
		})
	}

	var mdss []store.CephClusterMDS
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("filesystem asc, name asc").Find(&mdss).Error; err != nil {
		return detail, err
	}
	for _, mds := range mdss {
		detail.MDSs = append(detail.MDSs, cephDiscoveredRecord{
			Key:          mds.Name,
			Type:         mds.Filesystem,
			Hostname:     mds.Hostname,
			Status:       mds.State,
			Payload:      jsonPayload(mds.Payload),
			DiscoveredAt: mds.DiscoveredAt,
		})
	}

	var modules []store.CephClusterMgrModule
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&modules).Error; err != nil {
		return detail, err
	}
	for _, module := range modules {
		status := "disabled"
		if module.Enabled {
			status = "enabled"
		}
		detail.MgrModules = append(detail.MgrModules, cephDiscoveredRecord{
			Key:          module.Name,
			Status:       status,
			Payload:      jsonPayload(module.Payload),
			DiscoveredAt: module.DiscoveredAt,
		})
	}

	var configuration []store.CephClusterConfiguration
	if err := api.database().WithContext(ctx).Where("cluster_id = ?", clusterID).Order("who asc, name asc").Find(&configuration).Error; err != nil {
		return detail, err
	}
	for _, config := range configuration {
		detail.Configuration = append(detail.Configuration, cephDiscoveredRecord{
			Key:          strings.TrimSpace(config.Who + " " + config.Name),
			Type:         config.Level,
			Payload:      jsonPayload(config.Payload),
			DiscoveredAt: config.DiscoveredAt,
		})
	}

	return detail, nil
}

func jsonPayload(payload string) any {
	var decoded any
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return payload
	}
	return decoded
}

func toCephClusterResponse(cluster store.CephCluster) cephClusterResponse {
	return cephClusterResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: "",
		FSID:        "",
		Enabled:     true,
		Dashboard: dashboardConnectionResponse{
			Enabled:     true,
			BaseURL:     "",
			Username:    cluster.DashboardUsername,
			PasswordSet: cluster.DashboardPassword != "",
			InsecureTLS: false,
		},
		Command: commandConnectionResponse{
			Enabled:           true,
			Bin:               ceph.DefaultCephCommandBin,
			Cluster:           "",
			Conf:              "",
			MonitorHost:       cluster.MonitorHost,
			Name:              ceph.DefaultCephCommandName,
			Keyring:           "",
			KeyringContentSet: cluster.Keyring != "",
			TimeoutSeconds:    ceph.DefaultCephCommandTimeoutSeconds,
		},
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}

func (api *API) ClusterHealthFull(w http.ResponseWriter, r *http.Request) {
	health, err := api.ceph.HealthFull(r.Context())
	if err != nil {
		writeCephError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, health)
}
