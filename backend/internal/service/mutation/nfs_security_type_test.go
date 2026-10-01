package mutation

import (
	"encoding/json"
	"testing"
)

func TestNFSExportSecurityTypes(t *testing.T) {
	for _, value := range []any{[]string{}, []string{"none"}, []string{"sys"}, []string{"krb5", "krb5i", "krb5p"}} {
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "sectype": value}
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
				t.Fatal("security type payload mismatch")
			}
			record["sectype"] = []string{"sys"}
			data, err := nfsExportUpdateJSON(record, p)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(data, &record); err != nil {
				t.Fatal(err)
			}
			if !nfsExportAttributesMatch(record, p) {
				t.Fatal("security type update lost")
			}
			delete(record, "sectype")
			wantDefault := len(value.([]string)) == 0
			if nfsExportAttributesMatch(record, p) != wantDefault {
				t.Fatal("incorrect default readback")
			}
			record["sectype"] = []string{"invalid"}
			if nfsExportAttributesMatch(record, p) {
				t.Fatal("invalid readback accepted")
			}
		}
	}
	for _, value := range []any{nil, "sys", []string{"sys", "sys"}, []string{"invalid"}, []int{3}} {
		if _, err := nfsSecurityTypes(value); err == nil {
			t.Fatalf("accepted invalid security type %v", value)
		}
	}
	a, _ := nfsSecurityTypes([]string{"sys", "krb5"})
	b, _ := nfsSecurityTypes([]string{"krb5", "sys"})
	if a != b {
		t.Fatal("security type order matters")
	}
}
