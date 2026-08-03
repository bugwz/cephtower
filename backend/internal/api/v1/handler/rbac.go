package handler

import (
	"cephtower/backend/internal/store"
	"errors"
	"net/http"
	"strings"
	"time"
)

type createRoleRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type createRoleBindingRequest struct {
	ClusterID uint64 `json:"cluster_id"`
	UserID    uint64 `json:"user_id"`
	Role      string `json:"role"`
}

type deleteRoleBindingRequest struct {
	ClusterID uint64 `json:"cluster_id"`
	BindingID uint64 `json:"binding_id"`
}

type roleBindingDTO struct {
	ID        uint64    `json:"role_binding_id"`
	UserID    uint64    `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ClusterID uint64    `json:"cluster_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Database().ListRoles(r.Context())
	if err != nil {
		WriteError(w, r, 500, "store_error", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 200, "success", map[string]any{"items": rows, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]string{"request_id": RequestID(r)}})
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var request createRoleRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		WriteError(w, r, 400, "invalid_request", "name is required", false, nil)
		return
	}
	now := time.Now().UTC()
	role := store.Role{Name: request.Name, Description: request.Description, CreatedAt: now, UpdatedAt: now}
	if err := h.Database().CreateRole(r.Context(), &role); err != nil {
		WriteError(w, r, 400, "invalid_request", err.Error(), false, nil)
		return
	}
	WriteSuccess(w, 201, "success", role)
}

func (h *Handler) ListRoleBindings(w http.ResponseWriter, r *http.Request) {
	var request clusterIDRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	clusterID := request.ClusterID
	if clusterID == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id is required", false, nil)
		return
	}
	rows, err := h.Database().ListClusterRoleBindings(r.Context(), clusterID)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", "could not list role bindings", false, nil)
		return
	}
	items := make([]roleBindingDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, roleBindingDTO{ID: row.ID, UserID: row.UserID, Username: row.User.Username, Role: row.Role.Name, ClusterID: clusterID, CreatedAt: row.CreatedAt})
	}
	WriteSuccess(w, http.StatusOK, "success", map[string]any{"items": items, "pagination": map[string]any{"next_cursor": nil}, "meta": map[string]any{"request_id": RequestID(r)}})
}

func (h *Handler) CreateRoleBinding(w http.ResponseWriter, r *http.Request) {
	var request createRoleBindingRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	request.Role = strings.TrimSpace(request.Role)
	if request.ClusterID == 0 || request.UserID == 0 || request.Role == "" {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id, user_id and role are required", false, nil)
		return
	}
	clusterID := request.ClusterID
	var createdBy *uint64
	if actor, ok := CurrentUser(r); ok && actor.ID != 0 {
		createdBy = &actor.ID
	}
	if err := h.Database().BindUserRole(r.Context(), request.UserID, request.Role, &clusterID, createdBy); err != nil {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", err.Error(), false, nil)
		return
	}
	rows, err := h.Database().ListClusterRoleBindings(r.Context(), clusterID)
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", "could not read role binding", false, nil)
		return
	}
	for _, row := range rows {
		if row.UserID == request.UserID && row.Role.Name == request.Role {
			WriteSuccess(w, http.StatusCreated, "success", roleBindingDTO{ID: row.ID, UserID: row.UserID, Username: row.User.Username, Role: row.Role.Name, ClusterID: clusterID, CreatedAt: row.CreatedAt})
			return
		}
	}
	WriteError(w, r, http.StatusInternalServerError, "store_error", "role binding was not found after creation", false, nil)
}

func (h *Handler) DeleteRoleBinding(w http.ResponseWriter, r *http.Request) {
	var request deleteRoleBindingRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	if request.ClusterID == 0 || request.BindingID == 0 {
		WriteError(w, r, http.StatusBadRequest, "invalid_request", "cluster_id and binding_id are required", false, nil)
		return
	}
	err := h.Database().DeleteClusterRoleBinding(r.Context(), request.ClusterID, request.BindingID)
	if errors.Is(err, store.ErrRecordNotFound) {
		WriteError(w, r, http.StatusNotFound, "role_binding_not_found", "role binding was not found", false, nil)
		return
	}
	if err != nil {
		WriteError(w, r, http.StatusInternalServerError, "store_error", "could not delete role binding", false, nil)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", nil)
}
