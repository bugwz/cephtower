package mutation

import (
	"slices"
	"testing"
)

func TestSMBShareCreateReadonly(t *testing.T) {
	p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs"}
	for _, value := range []bool{true, false} {
		p["readonly"] = value
		spec, err := build(Request{Action: "smb_share.create"}, p)
		if err != nil || slices.Contains(spec.args, "--readonly") != value {
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
	if err != nil || slices.Contains(spec.args, "--readonly") {
		t.Fatalf("native default not preserved: %v %v", spec.args, err)
	}
}

func TestSMBShareCreateSubvolume(t *testing.T) {
	p := map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs"}
	for _, value := range []string{"docs", "team/docs", "_nogroup/docs"} {
		p["subvolume"] = value
		spec, err := build(Request{Action: "smb_share.create"}, p)
		if err != nil || !slices.Contains(spec.args, "--subvolume="+value) {
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
