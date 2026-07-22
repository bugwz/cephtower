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
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type Server struct {
	mu                sync.RWMutex
	cfg               config.Config
	ceph              v1.CephClient
	db                *gorm.DB
	syncCancel        context.CancelFunc
	clusterDiscoverer func(context.Context, *gorm.DB, *store.CephCluster) error
}

func NewServer(cfg config.Config, cephClient v1.CephClient, db *gorm.DB) *Server {
	server := &Server{
		cfg:               cfg,
		db:                db,
		clusterDiscoverer: discoverAndSyncCephCluster,
	}
	if cephClient == nil {
		cephClient = newDatabaseCephClient(server.database)
		server.startCephDataFetchScheduler()
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
	mux.HandleFunc("GET /healthz", s.healthz)
	s.registerAuthRoutes(mux)
	s.registerSetupRoutes(mux)
	s.registerClusterRoutes(mux)
	s.registerSystemConfigRoutes(mux)
	v1.RegisterRoutes(mux, s.ceph)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})
	mux.Handle("/", frontend.Handler())

	return withCORS(s.withAuth(mux))
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
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
		if r.Method == http.MethodOptions || r.URL.Path == "/healthz" || isPublicAPIPath(r.URL.Path) || !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		user, ok := s.userForRequest(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		if !canAccessPath(user, r.URL.Path) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "permission denied"})
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	})
}

func currentUser(r *http.Request) (store.User, bool) {
	user, ok := r.Context().Value(userContextKey{}).(store.User)
	return user, ok
}

type userContextKey struct{}

func isPublicAPIPath(path string) bool {
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/password-reset/request", "/api/v1/auth/password-reset/confirm", "/api/v1/setup/status", "/api/v1/setup/initialize":
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
	if action, ok := payload.(clusterActionResponse); ok && strings.TrimSpace(action.Message) != "" {
		return strings.TrimSpace(action.Message)
	}
	return fallback
}
