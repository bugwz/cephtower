package mutation

import (
	"encoding/json"
	"testing"
)

func TestSMBShareMaxConnections(t *testing.T) {
	p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs"}
	request := smbShareCreateRequest(p)
	before := []byte(`{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","max_connections":12,"cephfs":{"volume":"fs","path":"/"}}`)
	for _, value := range []any{float64(0), float64(25), json.Number("100")} {
		p["max_connections"] = value
		updated, err := smbShareUpdateJSON(before, request)
		if err != nil {
			t.Fatal(err)
		}
		if !smbShareUpdateMatches(updated, updated, request) || smbShareUpdateMatches(updated, before, request) {
			t.Fatal("incorrect connection limit readback")
		}
		created, err := smbShareCreateJSON(p, "a", "docs", "fs")
		if err != nil {
			t.Fatal(err)
		}
		var a, b map[string]any
		if json.Unmarshal(updated, &a) != nil || json.Unmarshal(created, &b) != nil || a["max_connections"] != b["max_connections"] {
			t.Fatal("creation lost connection limit")
		}
	}
	for _, value := range []any{float64(-1), 1.5, nil, false, "invalid"} {
		p["max_connections"] = value
		if _, err := smbShareUpdateJSON(before, request); err == nil {
			t.Fatalf("accepted invalid limit %v", value)
		}
	}
	delete(p, "max_connections")
	updated, err := smbShareUpdateJSON(before, request)
	if err != nil || !smbShareUpdateMatches(before, updated, request) {
		t.Fatal("omitted limit was not preserved", err)
	}
}
