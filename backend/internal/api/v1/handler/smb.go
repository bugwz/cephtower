package handler

import "net/http"

// SMB endpoints.
func (h *Handler) ListSMBClusters(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/cluster")
}

func (h *Handler) CreateSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/cluster")
}

func (h *Handler) GetSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/cluster/{cluster_id}")
}

func (h *Handler) DeleteSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/cluster/{cluster_id}")
}

func (h *Handler) ListSMBJoinAuths(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/joinauth")
}

func (h *Handler) CreateSMBJoinAuth(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/joinauth")
}

func (h *Handler) GetSMBJoinAuth(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/joinauth/{auth_id}")
}

func (h *Handler) DeleteSMBJoinAuth(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/joinauth/{auth_id}")
}

func (h *Handler) ListSMBShares(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/share")
}

func (h *Handler) CreateSMBShare(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/share")
}

func (h *Handler) GetSMBShare(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/share/{cluster_id}/{share_id}")
}

func (h *Handler) DeleteSMBShare(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/share/{cluster_id}/{share_id}")
}

func (h *Handler) ListSMBUsersGroups(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/usersgroups")
}

func (h *Handler) CreateSMBUsersGroups(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/usersgroups")
}

func (h *Handler) GetSMBUsersGroups(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/usersgroups/{users_groups_id}")
}

func (h *Handler) DeleteSMBUsersGroups(w http.ResponseWriter, r *http.Request) {
	h.serveDashboardPath(w, r, "/api/smb/usersgroups/{users_groups_id}")
}

type SMBClusterRequest = RawJSONRequest
type SMBClusterResponse = RawJSONResponse
type SMBJoinAuthRequest = RawJSONRequest
type SMBJoinAuthResponse = RawJSONResponse
type SMBShareRequest = RawJSONRequest
type SMBShareResponse = RawJSONResponse
type SMBUsersGroupsRequest = RawJSONRequest
type SMBUsersGroupsResponse = RawJSONResponse
