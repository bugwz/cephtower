package mutation

import (
	"encoding/json"
	"testing"
)

func TestNFSClientRules(t *testing.T) {
	for _, input := range []string{`[]`, `[{"addresses":["10.0.0.0/8","2001:db8::/64","*.example.com"],"access_type":"RO","squash":"root_squash"}]`, `[{"addresses":["*"],"access_type":"","squash":""}]`} {
		var clients any
		if err := json.Unmarshal([]byte(input), &clients); err != nil {
			t.Fatal(err)
		}
		p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "/", "filesystem": "cephfs", "client_rules": clients}
		cmd, err := build(Request{Action: "nfs_export.create"}, p)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(cmd.stdin, &record); err != nil {
			t.Fatal(err)
		}
		if !nfsExportAttributesMatch(record, p) {
			t.Fatal("client create payload mismatch")
		}
		record["clients"] = []any{}
		data, err := nfsExportUpdateJSON(record, p)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		if !nfsExportAttributesMatch(record, p) {
			t.Fatal("client update lost")
		}
		record["clients"] = nil
		if nfsExportAttributesMatch(record, p) {
			t.Fatal("missing client readback accepted")
		}
	}
	for _, input := range []string{`null`, `{}`, `[{}]`, `[{"addresses":[]}]`, `[{"addresses":["x;}"]}]`, `[{"addresses":["x\ny"]}]`, `[{"addresses":["host"],"access_type":"BAD"}]`, `[{"addresses":["host"],"squash":"BAD"}]`, `[{"addresses":["host"],"unknown":true}]`} {
		var value any
		if err := json.Unmarshal([]byte(input), &value); err != nil {
			t.Fatal(err)
		}
		if _, err := nfsClients(value); err == nil {
			t.Fatalf("invalid clients accepted: %s", input)
		}
	}
	var native any
	if err := json.Unmarshal([]byte(`[{"addresses":["*"],"access_type":null,"squash":null}]`), &native); err != nil {
		t.Fatal(err)
	}
	if !nfsClientsMatch(native, []nfsClientRule{{Addresses: []string{"*"}}}) {
		t.Fatal("inherited values must match")
	}
}
