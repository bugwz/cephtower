package mutation

import (
	"encoding/json"
	"testing"
)

func TestNFSMountRoot(t *testing.T) {
	for _, root := range []string{"/", "/data", "/data/./"} {
		p := map[string]any{"cluster": "nfs", "pseudo": "/share", "path": "/data/sub", "filesystem": "fs", "cmount_path": root}
		spec, err := build(Request{Action: "nfs_export.create"}, p)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(spec.stdin, &record); err != nil {
			t.Fatal(err)
		}
		if !nfsExportAttributesMatch(record, p) {
			t.Fatal("normalized mount readback failed")
		}
		record["fsal"].(map[string]any)["cmount_path"] = "/wrong"
		if nfsExportAttributesMatch(record, p) {
			t.Fatal("incorrect mount accepted")
		}
	}
	for _, root := range []any{"", "data", "/dat", "/data/sub/deeper", "/other", "/data/../../other", "/data\"", nil, 3} {
		p := map[string]any{"cluster": "nfs", "pseudo": "/share", "path": "/data/sub", "filesystem": "fs", "cmount_path": root}
		if _, err := build(Request{Action: "nfs_export.create"}, p); err == nil {
			t.Fatalf("accepted invalid mount %v", root)
		}
	}
	if _, err := nfsExportFSAL(map[string]any{"fsal_type": "RGW", "rgw_user_id": "owner", "cmount_path": "/"}); err == nil {
		t.Fatal("RGW mount accepted")
	}
}

func TestNFSMountUpdateIdentity(t *testing.T) {
	for _, changed := range []string{"none", "mount", "filesystem", "outside"} {
		fsal := map[string]any{"name": "CEPH", "fs_name": "fs", "cmount_path": "/data", "user_id": "old", "cephx_key": "old-key", "sec_label_xattr": "security.selinux"}
		record := map[string]any{"fsal": fsal, "pseudo": "/share", "path": "/data/sub"}
		p := map[string]any{"filesystem": "fs", "pseudo": "/share", "path": "/data/sub"}
		switch changed {
		case "mount":
			p["cmount_path"] = "/"
		case "filesystem":
			p["filesystem"] = "new"
		case "outside":
			p["path"] = "/elsewhere"
		}
		_, err := nfsExportUpdateJSON(record, p)
		if changed == "outside" {
			if err == nil {
				t.Fatal("retained mount containment not checked")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if changed == "none" {
			if fsal["user_id"] != "old" || fsal["cephx_key"] != "old-key" {
				t.Fatal("unchanged identity lost")
			}
		} else if fsal["user_id"] != nil || fsal["cephx_key"] != nil {
			t.Fatal("stale identity retained")
		}
		if fsal["sec_label_xattr"] != "security.selinux" {
			t.Fatal("unrelated attribute lost")
		}
	}
}
