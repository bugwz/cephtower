package router

import "cephtower/backend/internal/api/v1/handler"

func smbRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/smb/cluster", h.ListSMBClusters},
		{"POST", "/smb/cluster", h.CreateSMBCluster},
		{"GET", "/smb/cluster/{cluster_id}", h.GetSMBCluster},
		{"DELETE", "/smb/cluster/{cluster_id}", h.DeleteSMBCluster},
		{"GET", "/smb/joinauth", h.ListSMBJoinAuths},
		{"POST", "/smb/joinauth", h.CreateSMBJoinAuth},
		{"GET", "/smb/joinauth/{auth_id}", h.GetSMBJoinAuth},
		{"DELETE", "/smb/joinauth/{auth_id}", h.DeleteSMBJoinAuth},
		{"GET", "/smb/share", h.ListSMBShares},
		{"POST", "/smb/share", h.CreateSMBShare},
		{"GET", "/smb/share/{cluster_id}/{share_id}", h.GetSMBShare},
		{"DELETE", "/smb/share/{cluster_id}/{share_id}", h.DeleteSMBShare},
		{"GET", "/smb/usersgroups", h.ListSMBUsersGroups},
		{"POST", "/smb/usersgroups", h.CreateSMBUsersGroups},
		{"GET", "/smb/usersgroups/{users_groups_id}", h.GetSMBUsersGroups},
		{"DELETE", "/smb/usersgroups/{users_groups_id}", h.DeleteSMBUsersGroups},
	}
}
