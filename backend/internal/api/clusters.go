package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	Name              string `json:"name"`
	Keyring           string `json:"keyring"`
	KeyringContentSet bool   `json:"keyring_content_set"`
	TimeoutSeconds    int    `json:"timeout_seconds"`
}

type cephClusterRequest struct {
	Name              string `json:"name"`
	Keyring           string `json:"keyring"`
	DashboardUsername string `json:"dashboard_username"`
	DashboardPassword string `json:"dashboard_password"`
}

type clusterActionResponse struct {
	Message string `json:"message"`
}

type cephClusterDetailResponse struct {
	Cluster   cephClusterResponse         `json:"cluster"`
	Snapshots []cephResourceSnapshotEntry `json:"snapshots"`
}

type cephResourceSnapshotEntry struct {
	Category     string    `json:"category"`
	ResourceKey  string    `json:"resource_key"`
	Payload      any       `json:"payload"`
	LastSyncedAt time.Time `json:"last_synced_at"`
	LastError    string    `json:"last_error"`
}

func (s *Server) registerClusterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/clusters", s.listClusters)
	mux.HandleFunc("POST /api/v1/clusters", s.createCluster)
	mux.HandleFunc("GET /api/v1/clusters/{id}", s.getCluster)
	mux.HandleFunc("PUT /api/v1/clusters/{id}", s.updateCluster)
}

func (s *Server) listClusters(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	var clusters []store.CephCluster
	if err := s.database().Order("id asc").Find(&clusters).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := make([]cephClusterResponse, 0, len(clusters))
	for _, cluster := range clusters {
		response = append(response, toCephClusterResponse(cluster))
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) createCluster(w http.ResponseWriter, r *http.Request) {
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
	err = s.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cluster).Error; err != nil {
			return err
		}
		return s.discoverCephCluster(r.Context(), tx, &cluster)
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, clusterActionResponse{Message: "集群连接已创建"})
}

func (s *Server) getCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cluster, ok := s.clusterByID(w, r)
	if !ok {
		return
	}

	snapshots, err := s.clusterSnapshots(r.Context(), cluster.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cephClusterDetailResponse{
		Cluster:   toCephClusterResponse(cluster),
		Snapshots: snapshots,
	})
}

func (s *Server) updateCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	current, ok := s.clusterByID(w, r)
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

	err = s.database().WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&cluster).Error; err != nil {
			return err
		}
		return s.discoverCephCluster(r.Context(), tx, &cluster)
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, clusterActionResponse{Message: "集群连接已更新"})
}

func (s *Server) clusterByID(w http.ResponseWriter, r *http.Request) (store.CephCluster, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cluster id"})
		return store.CephCluster{}, false
	}

	var cluster store.CephCluster
	err = s.database().First(&cluster, id).Error
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
		Name:                  name,
		Enabled:               true,
		DashboardEnabled:      true,
		DashboardUsername:     strings.TrimSpace(req.DashboardUsername),
		CommandEnabled:        true,
		CommandBin:            "ceph",
		CommandName:           "client.admin",
		CommandTimeoutSeconds: 15,
	}

	if current != nil {
		cluster.ID = current.ID
		cluster.Description = current.Description
		cluster.FSID = current.FSID
		cluster.CreatedAt = current.CreatedAt
		cluster.DashboardBaseURL = current.DashboardBaseURL
		cluster.DashboardInsecureTLS = current.DashboardInsecureTLS
		cluster.CommandCluster = current.CommandCluster
		cluster.CommandConf = current.CommandConf
		cluster.CommandKeyring = current.CommandKeyring
		cluster.DashboardPassword = current.DashboardPassword
		cluster.CommandKeyringContent = current.CommandKeyringContent
	}
	if req.DashboardPassword != "" {
		cluster.DashboardPassword = req.DashboardPassword
	}
	if req.Keyring != "" {
		cluster.CommandKeyringContent = req.Keyring
	}

	if cluster.DashboardUsername == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard username is required")
	}
	if cluster.DashboardPassword == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard password is required")
	}
	if strings.TrimSpace(cluster.CommandKeyringContent) == "" {
		return store.CephCluster{}, fmt.Errorf("keyring is required")
	}

	return cluster, nil
}

func (s *Server) discoverCephCluster(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
	if s.clusterDiscoverer == nil {
		return nil
	}
	return s.clusterDiscoverer(ctx, db, cluster)
}

func (s *Server) clusterSnapshots(ctx context.Context, clusterID uint) ([]cephResourceSnapshotEntry, error) {
	var snapshots []store.CephResourceSnapshot
	if err := s.database().
		WithContext(ctx).
		Where("cluster_id = ?", clusterID).
		Order("category asc, resource_key asc").
		Find(&snapshots).Error; err != nil {
		return nil, err
	}

	response := make([]cephResourceSnapshotEntry, 0, len(snapshots))
	for _, snapshot := range snapshots {
		var payload any
		if err := json.Unmarshal([]byte(snapshot.Payload), &payload); err != nil {
			payload = snapshot.Payload
		}
		response = append(response, cephResourceSnapshotEntry{
			Category:     snapshot.Category,
			ResourceKey:  snapshot.ResourceKey,
			Payload:      payload,
			LastSyncedAt: snapshot.LastSyncedAt,
			LastError:    snapshot.LastError,
		})
	}
	return response, nil
}

func toCephClusterResponse(cluster store.CephCluster) cephClusterResponse {
	return cephClusterResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: cluster.Description,
		FSID:        cluster.FSID,
		Enabled:     cluster.Enabled,
		Dashboard: dashboardConnectionResponse{
			Enabled:     cluster.DashboardEnabled,
			BaseURL:     cluster.DashboardBaseURL,
			Username:    cluster.DashboardUsername,
			PasswordSet: cluster.DashboardPassword != "",
			InsecureTLS: cluster.DashboardInsecureTLS,
		},
		Command: commandConnectionResponse{
			Enabled:           cluster.CommandEnabled,
			Bin:               cluster.CommandBin,
			Cluster:           cluster.CommandCluster,
			Conf:              cluster.CommandConf,
			Name:              cluster.CommandName,
			Keyring:           cluster.CommandKeyring,
			KeyringContentSet: cluster.CommandKeyringContent != "",
			TimeoutSeconds:    cluster.CommandTimeoutSeconds,
		},
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
	}
}
