package mutation

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSMBLoginControlReplacement(t *testing.T) {
	r := Request{ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"filesystem": "fs"}}
	before := []byte(`{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","cephfs":{"volume":"fs","path":"/"},"login_control":[{"name":"old","category":"user","access":"read"}],"restrict_access":true}`)
	unchanged, err := smbShareUpdateJSON(before, r)
	if err != nil || !smbShareUpdateMatches(before, unchanged, r) {
		t.Fatal("omitted rules not preserved", err)
	}
	r.Parameters["login_control"] = []any{map[string]any{"name": "staff", "category": "group", "access": "read-write"}}
	r.Parameters["restrict_access"] = true
	updated, err := smbShareUpdateJSON(before, r)
	if err != nil || !smbShareUpdateMatches(updated, updated, r) || smbShareUpdateMatches(updated, before, r) {
		t.Fatal("replacement/readback failed", err)
	}
	r.Parameters["login_control"] = []any{}
	if _, err := smbShareUpdateJSON(before, r); err == nil {
		t.Fatal("empty allowlist accepted")
	}
	r.Parameters["restrict_access"] = false
	cleared, err := smbShareUpdateJSON(before, r)
	var record map[string]any
	if err != nil || json.Unmarshal(cleared, &record) != nil || len(record["login_control"].([]any)) != 0 {
		t.Fatal("clear failed", err)
	}
	delete(record, "restrict_access")
	native, _ := json.Marshal(record)
	if !smbShareUpdateMatches(cleared, native, r) {
		t.Fatal("native omitted false flag rejected")
	}
	for _, rules := range []any{nil, "bad", []any{nil}, []any{map[string]any{"name": "bad name", "category": "user", "access": "read"}}, []any{map[string]any{"name": "a", "category": "other", "access": "read"}}, []any{map[string]any{"name": "a", "category": "user", "access": "invalid"}}} {
		r.Parameters["login_control"] = rules
		if _, err := smbShareUpdateJSON(before, r); err == nil {
			t.Fatal("invalid rules accepted")
		}
	}
	r.Parameters["login_control"] = []any{map[string]any{"name": "a", "category": "user", "access": "none"}}
	r.Parameters["restrict_access"] = true
	if _, err := smbShareUpdateJSON(before, r); err == nil {
		t.Fatal("deny-only allowlist accepted")
	}
}
