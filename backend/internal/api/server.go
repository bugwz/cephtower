package api

import (
	"context"
	"encoding/json"
	"net/http"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/frontend"
	"cephtower/backend/internal/integrations/ceph"
)

type CephClient interface {
	ClusterSummary(ctx context.Context) (ceph.ClusterSummary, error)
}

type Server struct {
	cfg  config.Config
	ceph CephClient
}

func NewServer(cfg config.Config, cephClient CephClient) *Server {
	return &Server{
		cfg:  cfg,
		ceph: cephClient,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /api/v1/cluster/summary", s.clusterSummary)
	mux.HandleFunc("/api/", http.NotFound)
	mux.Handle("/", frontend.Handler())

	return withCORS(mux)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) clusterSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.ceph.ClusterSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}
