package router

import "cephtower/backend/internal/api/v1/handler"

func cephfsRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/filesystems", h.ListFilesystems},
		{"POST", "/filesystem", h.CreateFilesystem},
		{"GET", "/filesystem", h.GetFilesystem},
		{"PATCH", "/filesystem", h.UpdateFilesystem},
		{"DELETE", "/filesystem", h.DeleteFilesystem},
		{"GET", "/filesystem/clients", h.ListFilesystemClients},
		{"DELETE", "/filesystem/client", h.EvictFilesystemClient},
		{"GET", "/filesystem/subvolume/groups", h.ListSubvolumeGroups},
		{"POST", "/filesystem/subvolume/group", h.CreateSubvolumeGroup},
		{"GET", "/filesystem/subvolume/group", h.GetSubvolumeGroup},
		{"PATCH", "/filesystem/subvolume/group", h.UpdateSubvolumeGroup},
		{"DELETE", "/filesystem/subvolume/group", h.DeleteSubvolumeGroup},
		{"GET", "/filesystem/subvolumes", h.ListSubvolumes},
		{"POST", "/filesystem/subvolume", h.CreateSubvolume},
		{"GET", "/filesystem/subvolume", h.GetSubvolume},
		{"PATCH", "/filesystem/subvolume", h.UpdateSubvolume},
		{"DELETE", "/filesystem/subvolume", h.DeleteSubvolume},
		{"GET", "/filesystem/subvolume/snapshots", h.ListCephFSSnapshots},
		{"POST", "/filesystem/subvolume/snapshot", h.CreateCephFSSnapshot},
		{"POST", "/filesystem/subvolume/snapshot/clone", h.CloneCephFSSnapshot},
		{"GET", "/filesystem/snapshot/schedules", h.ListSnapshotSchedules},
		{"POST", "/filesystem/snapshot/schedule", h.CreateSnapshotSchedule},
		{"GET", "/filesystem/authorizations", h.ListCephFSAuthorizations},
		{"POST", "/filesystem/authorization", h.CreateCephFSAuthorization},
		{"GET", "/filesystem/entries", h.ListCephFSEntries},
		{"PATCH", "/filesystem/entry/quota", h.UpdateCephFSEntryQuota},
	}
}
