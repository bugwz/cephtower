package mutation

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNFSSecurityLabelAttribute(t *testing.T) {
	for _, label := range []string{"security.selinux", ""} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/data", "filesystem": "fs", "sec_label_xattr": label}
		for _, action := range []string{"nfs_export.create", "nfs_export.update"} {
			spec, err := build(Request{Action: action}, p)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(spec.stdin, &payload); err != nil {
				t.Fatal(err)
			}
			if payload["fsal"].(map[string]any)["sec_label_xattr"] != label {
				t.Fatal("attribute missing from apply payload")
			}
		}
		export := map[string]any{"pseudo": "/share", "path": "/data", "fsal": map[string]any{"name": "CEPH", "fs_name": "fs", "sec_label_xattr": "old", "user_id": "preserved"}}
		data, err := nfsExportUpdateJSON(export, p)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &export); err != nil {
			t.Fatal(err)
		}
		fsal := export["fsal"].(map[string]any)
		if fsal["sec_label_xattr"] != label || fsal["user_id"] != "preserved" {
			t.Fatal("update lost requested or existing fields")
		}
		if label == "" {
			delete(fsal, "sec_label_xattr")
		}
		if !nfsExportAttributesMatch(export, p) {
			t.Fatal("native readback rejected")
		}
		fsal["sec_label_xattr"] = "wrong"
		if nfsExportAttributesMatch(export, p) {
			t.Fatal("mismatched attribute accepted")
		}
		delete(p, "sec_label_xattr")
		if _, err := nfsExportUpdateJSON(export, p); err != nil || fsal["sec_label_xattr"] != "wrong" {
			t.Fatal("omission did not preserve existing attribute")
		}
	}
	for _, value := range []any{nil, 42, "x\"", "x\\", "x\n", "x y", strings.Repeat("a", 256)} {
		p := map[string]any{"filesystem": "fs", "path": "/", "sec_label_xattr": value}
		if _, err := nfsExportFSAL(p); err == nil {
			t.Fatalf("unsafe attribute accepted: %v", value)
		}
	}
	if _, err := nfsExportFSAL(map[string]any{"fsal_type": "RGW", "rgw_user_id": "owner", "sec_label_xattr": ""}); err == nil {
		t.Fatal("RGW accepted CephFS attribute")
	}
}
