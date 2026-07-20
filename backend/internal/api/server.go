package api

import (
	"encoding/json"
	"net/http"

	"cephtower/backend/internal/api/v1"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/frontend"
)

type Server struct {
	cfg  config.Config
	ceph v1.CephClient
}

func NewServer(cfg config.Config, cephClient v1.CephClient) *Server {
	return &Server{
		cfg:  cfg,
		ceph: cephClient,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	v1.RegisterRoutes(mux, s.ceph)
	mux.HandleFunc("/api/", http.NotFound)
	mux.Handle("/", frontend.Handler())

	return withCORS(mux)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
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
