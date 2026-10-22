package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/store"
)

type bindingsContextKey struct{}

// authorizeRoute runs before handlers, including handlers that decode their own bodies.
func (h *Handler) authorizeRoute(w http.ResponseWriter, r *http.Request, body []byte) (*http.Request, bool) {
	user, ok := CurrentUser(r)
	if !ok {
		WriteError(w, r, 401, "authentication_required", "authentication required", false, nil)
		return r, false
	}
	bindings, err := h.Database().ListUserRoleBindings(r.Context(), user.ID)
	if err != nil {
		WriteError(w, r, 500, "store_error", "could not verify permissions", false, nil)
		return r, false
	}
	r = r.WithContext(context.WithValue(r.Context(), bindingsContextKey{}, bindings))
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")
	if r.Method == http.MethodGet && (path == "/auth/me" || path == "/clusters") {
		return r, true
	}
	permission, global := routePermission(r.Method, path)
	var clusterID uint64
	if !global {
		var fields map[string]json.RawMessage
		valid := json.Unmarshal(body, &fields) == nil
		for key := range fields {
			if strings.EqualFold(key, "cluster_id") && key != "cluster_id" {
				valid = false
			}
		}
		// Typed DTOs and generic JSON handlers must authorize the same exact ID.
		// Generic handlers decode numbers as float64, so reject unsafe integers.
		if !valid || json.Unmarshal(fields["cluster_id"], &clusterID) != nil || clusterID == 0 || clusterID > 1<<53-1 {
			WriteError(w, r, 400, "invalid_request", "cluster_id must be a positive safe integer in the request body", false, nil)
			return r, false
		}
	}
	if !authservice.Allows(bindings, clusterID, permission) {
		WriteError(w, r, 403, "permission_denied", "permission denied for this action and cluster", false, nil)
		return r, false
	}
	return r, true
}

func routePermission(method, path string) (string, bool) {
	switch path {
	case "/user", "/role":
		return authservice.ManageSecurity, true
	case "/cluster":
		if method == http.MethodPost {
			return authservice.ManageCluster, true
		}
	case "/role/binding", "/role/bindings", "/credential", "/credentials", "/host/ssh", "/ceph/user", "/ceph/users", "/ceph/users/import", "/ceph/users/export", "/audit/events", "/filesystem/authorization", "/filesystem/authorizations":
		return authservice.ManageSecurity, false
	}
	if method == http.MethodGet {
		return authservice.ReadCluster, false
	}
	switch path {
	case "/resource/refresh", "/health/mute", "/alert/silence":
		return authservice.OperateCluster, false
	}
	for _, prefix := range []string{"/pool", "/rbd", "/filesystem", "/nfs", "/smb", "/rgw", "/iscsi", "/nvmeof"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return authservice.ManageStorage, false
		}
	}
	return authservice.ManageCluster, false
}

func (h *Handler) scopeClusterFilter(r *http.Request, filter store.ClusterFilter) store.ClusterFilter {
	if !h.RequireAuth() {
		return filter
	}
	bindings, _ := r.Context().Value(bindingsContextKey{}).([]store.UserRoleBinding)
	if authservice.Allows(bindings, 0, authservice.ReadCluster) {
		return filter
	}
	filter.IDs = []uint64{}
	for _, binding := range bindings {
		if binding.ClusterID != nil && authservice.Allows(bindings, *binding.ClusterID, authservice.ReadCluster) {
			filter.IDs = append(filter.IDs, *binding.ClusterID)
		}
	}
	return filter
}
