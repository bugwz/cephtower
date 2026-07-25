package handler

import (
	"net/http"
	"strings"
)

// WithAuth applies the authentication and authorization rules owned by the v1 contract.
func (h *Handler) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || isPublicAPIPath(r.URL.Path) || !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		user, ok := h.UserForRequest(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			return
		}
		if !CanAccessPath(user, r.URL.Path) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "permission denied"})
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), user)))
	})
}

func isPublicAPIPath(path string) bool {
	switch path {
	case "/api/v1/healthz", "/api/v1/readyz", "/api/v1/auth/login", "/api/v1/auth/password-reset/request", "/api/v1/auth/password-reset/confirm", "/api/v1/setup/status", "/api/v1/setup/database/test", "/api/v1/setup/initialize":
		return true
	default:
		return false
	}
}
