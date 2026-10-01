package mutation

import (
	"path"
	"regexp"
	"strings"
)

var nfsRGWUserID = regexp.MustCompile(`^[A-Za-z0-9_.:@$-]{1,512}$`)

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
		fsal := map[string]any{"name": "CEPH", "fs_name": filesystem}
		if value, exists := p["sec_label_xattr"]; exists {
			label, ok := value.(string)
			if !ok || len(label) > 255 || (label != "" && !regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(label)) {
				return nil, invalid("sec_label_xattr must be an attribute name of at most 255 bytes or empty to clear")
			}
			fsal["sec_label_xattr"] = label
		}
		return fsal, nil
	case "RGW":
		if _, exists := p["sec_label_xattr"]; exists {
			return nil, invalid("sec_label_xattr is only supported by CephFS exports")
		}
		if value, exists := p["rgw_bucket_tenant"]; exists {
			tenant, ok := value.(string)
			if !ok || (tenant != "" && (!regexp.MustCompile(`^[A-Za-z0-9_.-]{1,512}$`).MatchString(tenant) || strings.HasPrefix(tenant, "-"))) {
				return nil, invalid("invalid RGW bucket tenant")
			}
			bucket := optional(p, "path")
			if bucket == "" || strings.Contains(bucket, "/") || bucket == "." || bucket == ".." || strings.HasPrefix(bucket, "-") {
				return nil, invalid("invalid RGW bucket name")
			}
		}
		user := optional(p, "rgw_user_id")
		if user == "" {
			bucket := optional(p, "path")
			if bucket == "" || path.Clean(bucket) == "/" || path.Clean(bucket) == "." {
				return nil, invalid("an RGW bucket path or user id is required")
			}
			return map[string]any{"name": "RGW"}, nil
		}
		if !nfsRGWUserID.MatchString(user) || strings.HasPrefix(user, "-") {
			return nil, invalid("rgw_user_id is required or invalid")
		}
		return map[string]any{"name": "RGW", "user_id": user}, nil
	default:
		return nil, invalid("fsal_type must be CEPH or RGW")
	}
}
