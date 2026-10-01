package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func shareCreatePayload(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestSMBShareCreateAppliesFullConfiguration(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "smb_share.create", Parameters: map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs", "browseable": false, "comment": "private", "restrict_access": true, "login_control": []any{map[string]any{"name": "staff", "category": "group", "access": "read"}}}}
	spec, err := build(r, r.Parameters)
	if err != nil {
		t.Fatal(err)
	}
	native := `{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","name":"docs","intent":"present","readonly":false,"browseable":false,"comment":"private","cephfs":{"volume":"fs","path":"/","provider":"samba-vfs"},"restrict_access":true,"login_control":[{"name":"staff","category":"group","access":"read"}]}`
	if !smbShareUpdateMatches(spec.stdin, []byte(native), smbShareCreateRequest(r.Parameters)) {
		t.Fatal("native resource defaults not matched")
	}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action + ".pre_check": `[]`, r.Action: `{"success":true}`, r.Action + ".post_check": string(spec.stdin)}}
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if len(e.specs) != 3 || e.specs[0].Mutating || !e.specs[1].Mutating || e.specs[2].Mutating || strings.Join(e.specs[1].Args, " ") != "smb apply -i - --format json" {
		t.Fatal(e.specs)
	}
	if strings.Join(e.specs[2].Args, " ") != "smb show ceph.smb.share.a.docs --format json" {
		t.Fatal(e.specs[2])
	}
	e.outputs[r.Action+".pre_check"] = `["docs"]`
	e.specs = nil
	if _, err := s.Execute(context.Background(), r); err == nil || len(e.specs) != 1 {
		t.Fatal("existing share overwritten")
	}
	e.outputs[r.Action+".pre_check"] = `[]`
	for _, bad := range []string{`null`, `{}`, `["docs"]`, strings.Replace(string(spec.stdin), `"restrict_access":true`, `"restrict_access":false`, 1)} {
		e.outputs[r.Action+".post_check"] = bad
		if _, err := s.Execute(context.Background(), r); err == nil {
			t.Fatal("invalid creation readback accepted")
		}
	}
}

func TestSMBShareCreateReadonly(t *testing.T) {
	p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs"}
	for _, value := range []bool{true, false} {
		p["readonly"] = value
		spec, err := build(Request{Action: "smb_share.create"}, p)
		if err != nil || shareCreatePayload(t, spec.stdin)["readonly"] != value {
			t.Fatalf("readonly=%v: %v %v", value, spec.args, err)
		}
	}
	for _, value := range []any{"false", 0, nil} {
		p["readonly"] = value
		if _, err := build(Request{Action: "smb_share.create"}, p); err == nil {
			t.Fatalf("invalid readonly accepted: %v", value)
		}
	}
	delete(p, "readonly")
	spec, err := build(Request{Action: "smb_share.create"}, p)
	if err != nil || shareCreatePayload(t, spec.stdin)["readonly"] != false {
		t.Fatalf("native default not preserved: %v %v", spec.args, err)
	}
}

func TestSMBShareCreateSubvolume(t *testing.T) {
	p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs"}
	for _, value := range []string{"docs", "team/docs", "_nogroup/docs"} {
		p["subvolume"] = value
		spec, err := build(Request{Action: "smb_share.create"}, p)
		if err != nil {
			t.Fatal(err)
		}
		storage := shareCreatePayload(t, spec.stdin)["cephfs"].(map[string]any)
		parts := strings.Split(value, "/")
		if storage["subvolume"] != parts[len(parts)-1] || (len(parts) == 2 && storage["subvolumegroup"] != parts[0]) {
			t.Fatalf("subvolume=%s: %v %v", value, spec.args, err)
		}
	}
	for _, value := range []any{"", "a/b/c", "/a", "a/", "../a", "a/..", "a\n", "a\x00", false} {
		p["subvolume"] = value
		if _, err := build(Request{Action: "smb_share.create"}, p); err == nil {
			t.Fatalf("invalid subvolume accepted: %v", value)
		}
	}
}
