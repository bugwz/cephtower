package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"cephtower/backend/internal/store"
)

type userContextKey struct{}

func (h *Handler) UserForRequest(r *http.Request) (store.User, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" || token == r.Header.Get("Authorization") {
		return store.User{}, false
	}

	user, err := h.auth.UserForToken(r.Context(), token)
	if err != nil {
		return store.User{}, false
	}
	return user, true
}

func ContextWithUser(ctx context.Context, user store.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func currentUser(r *http.Request) (store.User, bool) {
	user, ok := r.Context().Value(userContextKey{}).(store.User)
	return user, ok
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	user, ok := currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return false
	}
	if user.Role != store.UserRoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "administrator role required"})
		return false
	}
	return true
}

func CanAccessPath(user store.User, path string) bool {
	if user.Role == store.UserRoleAdmin {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/auth/") {
		return true
	}
	if path == "/api/v1/user" || strings.HasPrefix(path, "/api/v1/user/") {
		return false
	}
	if isClusterManagementPath(path) {
		return false
	}

	switch {
	case strings.Contains(path, "/configuration"), strings.Contains(path, "/log"):
		return hasPermission(user, "system:read")
	case strings.Contains(path, "/storage"), strings.Contains(path, "/pool"), strings.Contains(path, "/block"), strings.Contains(path, "/filesystem"), strings.Contains(path, "/object"):
		return hasPermission(user, "storage:read")
	default:
		return hasPermission(user, "cluster:read")
	}
}

func isClusterManagementPath(path string) bool {
	if path == "/api/v1/cluster" {
		return true
	}
	if !strings.HasPrefix(path, "/api/v1/cluster/") {
		return false
	}
	segment := strings.TrimPrefix(path, "/api/v1/cluster/")
	if index := strings.IndexByte(segment, '/'); index >= 0 {
		segment = segment[:index]
	}
	_, err := strconv.ParseUint(segment, 10, 64)
	return err == nil
}

func hasPermission(user store.User, permission string) bool {
	var permissions []string
	if err := json.Unmarshal([]byte(user.Permissions), &permissions); err != nil {
		return false
	}
	for _, item := range permissions {
		if item == permission {
			return true
		}
	}
	return false
}
