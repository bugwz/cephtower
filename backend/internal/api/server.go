package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"cephtower/backend/internal/api/v1"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/frontend"
	"cephtower/backend/internal/service/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type Server struct {
	mu                sync.RWMutex
	cfg               config.Config
	ceph              v1.CephClient
	db                *gorm.DB
	syncCancel        context.CancelFunc
	clusterDiscoverer ceph.ClusterDiscoverer
}

func NewServer(cfg config.Config, cephClient v1.CephClient, db *gorm.DB) *Server {
	runtimeDir := config.ResolveRuntimeDir(cfg)
	server := &Server{
		cfg: cfg,
		db:  db,
		clusterDiscoverer: func(ctx context.Context, db *gorm.DB, cluster *store.CephCluster) error {
			return ceph.DiscoverAndSyncCephClusterWithWorkDir(ctx, db, cluster, runtimeDir)
		},
	}
	if cephClient == nil {
		cephClient = ceph.NewDatabaseCephClient(server.database, runtimeDir)
		server.syncCancel = ceph.StartDataFetchScheduler(server.database, runtimeDir)
	}
	server.ceph = cephClient
	return server
}

func (s *Server) Close() error {
	if s.syncCancel != nil {
		s.syncCancel()
	}
	return store.Close(s.database())
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	s.registerAPIRouter(mux)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})
	mux.Handle("/", frontend.Handler())

	return withCORS(s.withAuth(mux))
}

func (s *Server) currentConfig() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Server) database() *gorm.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}

func (s *Server) replaceDatabase(cfg config.Config, db *gorm.DB) *gorm.DB {
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.db
	s.cfg = cfg
	s.db = db
	return previous
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || isPublicAPIPath(r.URL.Path) || !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		user, ok := v1.UserForRequest(s.database, r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		if !v1.CanAccessPath(user, r.URL.Path) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "permission denied"})
			return
		}

		next.ServeHTTP(w, r.WithContext(v1.ContextWithUser(r.Context(), user)))
	})
}

func isPublicAPIPath(path string) bool {
	switch path {
	case v1.PathPrefix + "/healthz",
		v1.PathPrefix + "/auth/login",
		v1.PathPrefix + "/auth/password-reset/request",
		v1.PathPrefix + "/auth/password-reset/confirm",
		v1.PathPrefix + "/setup/status",
		v1.PathPrefix + "/setup/initialize":
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusForAPIResponse(status))
	_ = json.NewEncoder(w).Encode(apiResponseForStatus(status, payload))
}

func httpStatusForAPIResponse(status int) int {
	if status == http.StatusUnauthorized || status == http.StatusForbidden || status >= http.StatusInternalServerError {
		return status
	}
	return http.StatusOK
}

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func apiResponseForStatus(status int, payload any) apiResponse {
	if status >= http.StatusBadRequest {
		return apiResponse{
			Code:    status,
			Message: responseMessage(payload, http.StatusText(status)),
			Data:    nil,
		}
	}

	return apiResponse{
		Code:    0,
		Message: responseMessage(payload, "success"),
		Data:    payload,
	}
}

func responseMessage(payload any, fallback string) string {
	if values, ok := payload.(map[string]string); ok {
		for _, key := range []string{"message", "error"} {
			if message := strings.TrimSpace(values[key]); message != "" {
				return message
			}
		}
	}
	if values, ok := payload.(map[string]any); ok {
		for _, key := range []string{"message", "error"} {
			if message, ok := values[key].(string); ok && strings.TrimSpace(message) != "" {
				return strings.TrimSpace(message)
			}
		}
	}
	return fallback
}
