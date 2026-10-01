package mutation

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSMBShareNativePaths(t *testing.T) {
	for input, want := range map[string]string{"/": "/", "/data//docs/": "/data/docs", "/data/../docs": "/docs", "docs/./team": "docs/team", "//server/data/": "//server/data", "///data": "/data", "/../../docs": "/docs"} {
		got, err := smbSharePath(input)
		if err != nil || got != want {
			t.Fatalf("%q: %q %v", input, got, err)
		}
		p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs", "path": input}
		spec, err := build(Request{Action: "smb_share.create"}, p)
		if err != nil || shareCreatePayload(t, spec.stdin)["cephfs"].(map[string]any)["path"] != want {
			t.Fatalf("create path: %v %v", spec.args, err)
		}
		request := Request{ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: p}
		before := []byte(`{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","cephfs":{"volume":"fs","path":"/old"}}`)
		data, err := smbShareUpdateJSON(before, request)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(data, &record); err != nil || record["cephfs"].(map[string]any)["path"] != want {
			t.Fatalf("update path: %s %v", data, err)
		}
		if !smbShareUpdateMatches(data, data, request) {
			t.Fatal("normalized readback rejected")
		}
	}
	for _, input := range []string{"", ".", "..", "../docs", "a/../../docs", "//", "a\nb", "a\x00b"} {
		if _, err := smbSharePath(input); err == nil {
			t.Fatalf("invalid path accepted: %q", input)
		}
	}
}
