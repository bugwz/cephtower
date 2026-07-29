package api

import (
	"net/http"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
	"cephtower/backend/internal/webui"
)

type API struct {
	handler *handler.Handler
}

// NewAPI builds the HTTP transport from already-constructed application
// services. Dependency construction and resource ownership belong to app.
func NewAPI(apiHandler *handler.Handler) *API {
	return &API{handler: apiHandler}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	router.Register(mux, a.handler)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		handler.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	})
	mux.Handle("/", webui.WebUIHandler())
	return withCORS(withMethodOverride(mux))
}
