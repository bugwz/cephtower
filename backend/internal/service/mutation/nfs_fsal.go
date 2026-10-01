package mutation

import "strings"

func nfsExportFSAL(p map[string]any) (map[string]any, error) {
	kind := optional(p, "fsal_type")
	if kind == "" {
		kind = "CEPH"
	}
	switch kind {
	case "CEPH":
		filesystem, err := required(p, "filesystem")
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(optional(p, "path"), "/") {
			return nil, invalid("CephFS export path must be absolute")
		}
		return map[string]any{"name": "CEPH", "fs_name": filesystem}, nil
	case "RGW":
		user, err := required(p, "rgw_user_id")
		if err != nil {
			return nil, err
		}
		return map[string]any{"name": "RGW", "user_id": user}, nil
	default:
		return nil, invalid("fsal_type must be CEPH or RGW")
	}
}
