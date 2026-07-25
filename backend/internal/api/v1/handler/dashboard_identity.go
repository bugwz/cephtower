package handler

import "net/http"

func (h *Handler) ListDashboardUsers(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user")
}

func (h *Handler) CreateDashboardUser(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user")
}

func (h *Handler) ValidateDashboardUserPassword(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user/validate_password")
}

func (h *Handler) GetDashboardUser(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user/{username}")
}

func (h *Handler) UpdateDashboardUser(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user/{username}")
}

func (h *Handler) DeleteDashboardUser(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user/{username}")
}

func (h *Handler) ChangeDashboardUserPassword(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/user/{username}/change_password")
}

func (h *Handler) ListDashboardRoles(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/role")
}

func (h *Handler) CreateDashboardRole(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/role")
}

func (h *Handler) GetDashboardRole(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/role/{name}")
}

func (h *Handler) UpdateDashboardRole(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/role/{name}")
}

func (h *Handler) DeleteDashboardRole(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/role/{name}")
}

func (h *Handler) CloneDashboardRole(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/role/{name}/clone")
}

type DashboardUserRequest = RawJSONRequest
type DashboardUserResponse = RawJSONResponse
type DashboardRoleRequest = RawJSONRequest
type DashboardRoleResponse = RawJSONResponse
type DashboardPasswordValidationRequest = RawJSONRequest
type DashboardPasswordValidationResponse = RawJSONResponse
