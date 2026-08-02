package store

import "strings"

var entityKinds = []string{
	"capability",
	"cephfs_authorization",
	"cephfs_client",
	"cephfs_entry",
	"cephfs_snapshot",
	"config_option",
	"config_value",
	"crush_rule",
	"daemon",
	"device",
	"erasure_code_profile",
	"filesystem",
	"health_check",
	"log",
	"mds",
	"mgr",
	"mgr_module",
	"mon",
	"nfs_cluster",
	"nfs_export",
	"osd",
	"osd_flag",
	"osd_removal",
	"pool",
	"rbd_group",
	"rbd_image",
	"rbd_mirroring",
	"rbd_namespace",
	"rbd_snapshot",
	"rbd_trash",
	"rgw_account",
	"rgw_bucket",
	"rgw_realm",
	"rgw_role",
	"rgw_status",
	"rgw_user",
	"rgw_zone",
	"rgw_zonegroup",
	"service",
	"smb_cluster",
	"smb_share",
	"snapshot_schedule",
	"subvolume",
	"subvolume_group",
	"upgrade",
}

var entityKindSet = func() map[string]struct{} {
	result := make(map[string]struct{}, len(entityKinds))
	for _, kind := range entityKinds {
		result[kind] = struct{}{}
	}
	return result
}()

func EntityKinds() []string {
	return append([]string(nil), entityKinds...)
}

func EntityTableName(kind string) (string, bool) {
	kind = strings.TrimSpace(kind)
	_, ok := entityKindSet[kind]
	if !ok {
		return "", false
	}
	return "ceph_" + kind, true
}
