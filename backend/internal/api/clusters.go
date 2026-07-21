package api

import (
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

type dashboardConnectionRequest struct {
	Enabled     bool   `json:"enabled"`
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	ClearSecret bool   `json:"clear_secret"`
	InsecureTLS bool   `json:"insecure_tls"`
}

type dashboardConnectionResponse struct {
	Enabled     bool   `json:"enabled"`
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	PasswordSet bool   `json:"password_set"`
	InsecureTLS bool   `json:"insecure_tls"`
}

type commandConnectionRequest struct {
	Enabled        bool   `json:"enabled"`
	Bin            string `json:"bin"`
	Cluster        string `json:"cluster"`
	Conf           string `json:"conf"`
	Name           string `json:"name"`
	Keyring        string `json:"keyring"`
	KeyringContent string `json:"keyring_content"`
	ClearSecret    bool   `json:"clear_secret"`
	TimeoutSeconds int    `json:"timeout_seconds"`
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
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	FSID        string                     `json:"fsid"`
	Enabled     *bool                      `json:"enabled"`
	Dashboard   dashboardConnectionRequest `json:"dashboard"`
	Command     commandConnectionRequest   `json:"command"`
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
	if err := s.database().Create(&cluster).Error; err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, toCephClusterResponse(cluster))
}

func (s *Server) getCluster(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cluster, ok := s.clusterByID(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, toCephClusterResponse(cluster))
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

	if err := s.database().Save(&cluster).Error; err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, toCephClusterResponse(cluster))
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
	if !req.Dashboard.Enabled && !req.Command.Enabled {
		return store.CephCluster{}, fmt.Errorf("at least one ceph access method must be enabled")
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	cluster := store.CephCluster{
		Name:                  name,
		Description:           strings.TrimSpace(req.Description),
		FSID:                  strings.TrimSpace(req.FSID),
		Enabled:               enabled,
		DashboardEnabled:      req.Dashboard.Enabled,
		DashboardBaseURL:      strings.TrimRight(strings.TrimSpace(req.Dashboard.BaseURL), "/"),
		DashboardUsername:     strings.TrimSpace(req.Dashboard.Username),
		DashboardInsecureTLS:  req.Dashboard.InsecureTLS,
		CommandEnabled:        req.Command.Enabled,
		CommandBin:            strings.TrimSpace(req.Command.Bin),
		CommandCluster:        strings.TrimSpace(req.Command.Cluster),
		CommandConf:           strings.TrimSpace(req.Command.Conf),
		CommandName:           strings.TrimSpace(req.Command.Name),
		CommandKeyring:        strings.TrimSpace(req.Command.Keyring),
		CommandTimeoutSeconds: req.Command.TimeoutSeconds,
	}
	if cluster.CommandBin == "" {
		cluster.CommandBin = "ceph"
	}
	if cluster.CommandTimeoutSeconds == 0 {
		cluster.CommandTimeoutSeconds = 15
	}
	if cluster.CommandTimeoutSeconds < 0 {
		return store.CephCluster{}, fmt.Errorf("command timeout_seconds must be greater than zero")
	}
	if cluster.DashboardEnabled && cluster.DashboardBaseURL == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard base_url is required when dashboard access is enabled")
	}

	if current != nil {
		cluster.DashboardPassword = current.DashboardPassword
		cluster.CommandKeyringContent = current.CommandKeyringContent
	}
	if req.Dashboard.ClearSecret {
		cluster.DashboardPassword = ""
	} else if req.Dashboard.Password != "" {
		cluster.DashboardPassword = req.Dashboard.Password
	}
	if req.Command.ClearSecret {
		cluster.CommandKeyringContent = ""
	} else if req.Command.KeyringContent != "" {
		cluster.CommandKeyringContent = req.Command.KeyringContent
	}

	return cluster, nil
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
