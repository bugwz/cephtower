package mutation

import (
	"encoding/json"
	"testing"
)

func TestNFSExportTransports(t *testing.T) {
	for _, value := range []any{[]string{"TCP"}, []string{"UDP"}, []string{"UDP", "TCP"}} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "transports": value}
		for _, action := range []string{"nfs_export.create", "nfs_export.update"} {
			cmd, err := build(Request{Action: action}, p)
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]any
			if err := json.Unmarshal(cmd.stdin, &record); err != nil {
				t.Fatal(err)
			}
			if !nfsExportAttributesMatch(record, p) {
				t.Fatal("transport payload mismatch")
			}
			record["access_type"] = "RW"
			record["transports"] = []string{"TCP", "UDP"}
			data, err := nfsExportUpdateJSON(record, p)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &record); err != nil {
				t.Fatal(err)
			}
			if !nfsExportAttributesMatch(record, p) {
				t.Fatal("update transports lost")
			}
			record["transports"] = nil
			if nfsExportAttributesMatch(record, p) {
				t.Fatal("missing readback accepted")
			}
		}
	}
	for _, value := range []any{nil, []string{}, []string{"TCP", "TCP"}, []string{"tcp"}, []string{"SCTP"}, "TCP", []int{3}, []any{true}} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "transports": value}
		if _, err := build(Request{Action: "nfs_export.create"}, p); err == nil {
			t.Fatalf("invalid transports accepted: %v", value)
		}
	}
	left, err := nfsTransports([]string{"TCP", "UDP"})
	right, otherErr := nfsTransports([]string{"UDP", "TCP"})
	if err != nil || otherErr != nil || left != right {
		t.Fatal("transport order must not matter")
	}
}
