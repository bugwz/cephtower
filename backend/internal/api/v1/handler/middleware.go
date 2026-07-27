package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

func (h *Handler) PrepareRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			b := make([]byte, 12)
			_, _ = rand.Read(b)
			requestID = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := WithRequestID(r.Context(), requestID)
		if r.URL.Query().Has("cluster_id") || r.Header.Get("X-Cluster-ID") != "" {
			r = r.WithContext(ctx)
			WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id must use the request body", false, nil)
			return
		}
		if cookie, err := r.Cookie("cluster_id"); err == nil && cookie.Value != "" {
			r = r.WithContext(ctx)
			WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id must use the request body", false, nil)
			return
		}
		r = r.WithContext(ctx)
		if isPublicRoute(r) {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" || token == r.Header.Get("Authorization") {
			WriteError(w, r, http.StatusUnauthorized, "authentication_required", "authentication required", false, nil)
			return
		}
		user, err := h.Auth.UserForToken(r.Context(), token)
		if err != nil {
			WriteError(w, r, http.StatusUnauthorized, "authentication_required", "authentication required", false, nil)
			return
		}
		ctx = WithUser(r.Context(), user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func isPublicRoute(r *http.Request) bool {
	switch r.Method + " " + r.URL.Path {
	case "GET /api/v1/healthz",
		"GET /api/v1/readyz",
		"GET /api/v1/bootstrap",
		"POST /api/v1/bootstrap/admin",
		"POST /api/v1/auth/login":
		return true
	default:
		return false
	}
}
