package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) PrepareRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			b := make([]byte, 12)
			_, _ = rand.Read(b)
			requestID = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", requestID)
		body := readAuditBody(r)
		info := &auditInfo{Action: r.Method + " " + strings.TrimPrefix(r.URL.Path, "/api/v1")}
		ctx := withAuditInfo(WithRequestID(r.Context(), requestID), info)
		if r.URL.Query().Has("cluster_id") || r.Header.Get("X-Cluster-ID") != "" {
			r = r.WithContext(ctx)
			recorder := &auditResponseRecorder{ResponseWriter: w}
			WriteError(recorder, r, http.StatusBadRequest, "invalid_request", "cluster_id must use the request body", false, nil)
			h.recordAuditEvent(r, recorder, body, started)
			return
		}
		if cookie, err := r.Cookie("cluster_id"); err == nil && cookie.Value != "" {
			r = r.WithContext(ctx)
			recorder := &auditResponseRecorder{ResponseWriter: w}
			WriteError(recorder, r, http.StatusBadRequest, "invalid_request", "cluster_id must use the request body", false, nil)
			h.recordAuditEvent(r, recorder, body, started)
			return
		}
		r = r.WithContext(ctx)
		if isPublicRoute(r) {
			h.auditRequest(next, w, r, body, started)
			return
		}
		if h.Database == nil || h.Database() == nil {
			recorder := &auditResponseRecorder{ResponseWriter: w}
			WriteError(recorder, r, http.StatusServiceUnavailable, "not_ready", "database is unavailable", true, nil)
			h.recordAuditEvent(r, recorder, body, started)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" || token == r.Header.Get("Authorization") {
			recorder := &auditResponseRecorder{ResponseWriter: w}
			WriteError(recorder, r, http.StatusUnauthorized, "authentication_required", "authentication required", false, nil)
			h.recordAuditEvent(r, recorder, body, started)
			return
		}
		user, err := h.Auth.UserForToken(r.Context(), token)
		if err != nil {
			recorder := &auditResponseRecorder{ResponseWriter: w}
			WriteError(recorder, r, http.StatusUnauthorized, "authentication_required", "authentication required", false, nil)
			h.recordAuditEvent(r, recorder, body, started)
			return
		}
		ctx = WithUser(r.Context(), user)
		r = r.WithContext(ctx)
		h.auditRequest(next, w, r, body, started)
	})
}

func readAuditBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, auditMaxBody))
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

func isPublicRoute(r *http.Request) bool {
	switch r.Method + " " + r.URL.Path {
	case "GET /api/v1/healthz",
		"GET /api/v1/readyz",
		"GET /api/v1/bootstrap",
		"POST /api/v1/bootstrap/dbtest",
		"POST /api/v1/bootstrap/run",
		"POST /api/v1/auth/login":
		return true
	default:
		return false
	}
}
