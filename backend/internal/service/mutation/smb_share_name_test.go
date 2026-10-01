package mutation

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestSMBShareComment(t *testing.T) {
	request := Request{ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"filesystem": "fs"}}
	before := []byte(`{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","comment":"old","cephfs":{"volume":"fs","path":"/"}}`)
	data, err := smbShareUpdateJSON(before, request)
	if err != nil || !smbShareUpdateMatches(before, data, request) {
		t.Fatal("omitted comment not preserved", err)
	}
	for _, comment := range []string{"团队资料", ""} {
		request.Parameters["comment"] = comment
		data, err := smbShareUpdateJSON(before, request)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(data, &record); err != nil || record["comment"] != comment {
			t.Fatalf("wrong comment: %s %v", data, err)
		}
		if !smbShareUpdateMatches(data, data, request) || smbShareUpdateMatches(data, before, request) {
			t.Fatal("comment readback failed")
		}
	}
	for _, bad := range []any{nil, true, 1, "a\nb", "a\rb", "a\x00b"} {
		request.Parameters["comment"] = bad
		if _, err := smbShareUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid comment accepted: %v", bad)
		}
	}
}

func TestSMBShareDisplayName(t *testing.T) {
	p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs", "share_name": "Team Documents"}
	spec, err := build(Request{Action: "smb_share.create"}, p)
	if err != nil || shareCreatePayload(t, spec.stdin)["name"] != "Team Documents" || shareCreatePayload(t, spec.stdin)["share_id"] != "docs" {
		t.Fatalf("name and identity not separated: %v %v", spec.args, err)
	}
	request := Request{ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: p}
	before := []byte(`{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","name":"Old Name","cephfs":{"volume":"fs","path":"/"}}`)
	data, err := smbShareUpdateJSON(before, request)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil || record["name"] != "Team Documents" || record["share_id"] != "docs" {
		t.Fatalf("invalid renamed resource: %s %v", data, err)
	}
	if smbShareUpdateMatches(data, before, request) || !smbShareUpdateMatches(data, data, request) {
		t.Fatal("name readback not verified")
	}
	for _, bad := range []any{"", " leading", "bad/name", "bad\n", "中文", strings.Repeat("a", 65), false, nil} {
		p["share_name"] = bad
		if _, err := build(Request{Action: "smb_share.create"}, p); err == nil {
			t.Fatalf("create accepted invalid name: %v", bad)
		}
		if _, err := smbShareUpdateJSON(before, request); err == nil {
			t.Fatalf("update accepted invalid name: %v", bad)
		}
	}
	for _, good := range []string{"_", "A.b _-", strings.Repeat("a", 64)} {
		p["share_name"] = good
		if _, _, err := smbShareName(p); err != nil {
			t.Fatal(err)
		}
	}
	delete(p, "share_name")
	data, err = smbShareUpdateJSON(before, request)
	if err != nil || !smbShareUpdateMatches(before, data, request) {
		t.Fatalf("omitted name not preserved: %s %v", data, err)
	}
}
