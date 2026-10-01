package mutation

import (
	"encoding/json"
	"testing"
)

func TestNFSExportAccessTypes(t *testing.T) {
	for _, access := range []string{"RO", "RW", "NONE"} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "access_type": access}
		for _, action := range []string{"nfs_export.create", "nfs_export.update"} {
			cmd, err := build(Request{Action: action}, p)
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]any
			if err := json.Unmarshal(cmd.stdin, &record); err != nil {
				t.Fatal(err)
			}
			if record["access_type"] != access || !nfsExportAttributesMatch(record, p) {
				t.Fatal("access payload mismatch")
			}
			record["access_type"] = "NONE"
			record["clients"] = []any{map[string]any{"addresses": []string{"10.0.0.0/8"}, "access_type": "RW"}}
			data, err := nfsExportUpdateJSON(record, p)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &record); err != nil {
				t.Fatal(err)
			}
			if !nfsExportAttributesMatch(record, p) || len(record["clients"].([]any)) != 1 {
				t.Fatal("access update lost attributes")
			}
			record["access_type"] = "invalid"
			if nfsExportAttributesMatch(record, p) {
				t.Fatal("access mismatch accepted")
			}
		}
	}
	for _, value := range []any{nil, true, "", "INVALID", "ro"} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "access_type": value}
		if _, err := build(Request{Action: "nfs_export.create"}, p); err == nil {
			t.Fatalf("invalid access accepted: %v", value)
		}
	}
}
