package api

import (
	"net/http"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
	"cephtower/backend/internal/webui"
)

type Server struct {
	handler *handler.Handler
}

// NewServer builds the HTTP transport from already-constructed application
// services. Dependency construction and resource ownership belong to app.
func NewServer(apiHandler *handler.Handler) *Server {
	return &Server{handler: apiHandler}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	router.Register(mux, s.handler)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		handler.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})
	mux.Handle("/", webui.WebUIHandler())
	return withCORS(s.handler.WithAuth(mux))
}
