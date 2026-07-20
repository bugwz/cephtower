package httpapi

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"cephtower/backend/internal/ceph"
	"cephtower/backend/internal/config"
)

//go:embed static/dist
var staticFiles embed.FS

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
	mux.Handle("/", frontendHandler())

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

func frontendHandler() http.Handler {
	dist, err := fs.Sub(staticFiles, "static/dist")
	if err != nil {
		return http.NotFoundHandler()
	}

	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		if staticFileExists(dist, r.URL.Path) {
			fileServer.ServeHTTP(w, r)
			return
		}

		indexRequest := new(http.Request)
		*indexRequest = *r
		indexURL := *r.URL
		indexURL.Path = "/"
		indexRequest.URL = &indexURL
		fileServer.ServeHTTP(w, indexRequest)
	})
}

func staticFileExists(dist fs.FS, requestPath string) bool {
	name := strings.TrimPrefix(path.Clean(requestPath), "/")
	if name == "." || name == "" {
		return true
	}

	file, err := dist.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	return err == nil && !stat.IsDir()
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
