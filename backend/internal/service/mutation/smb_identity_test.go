package mutation

import (
	"strings"
	"testing"
)

func TestSMBNativeCreationIDs(t *testing.T) {
	for _, action := range []string{"smb_cluster.create", "smb_share.create"} {
		for _, name := range []string{"a", "A-1", strings.Repeat("a", 18)} {
			p := map[string]any{"name": name, "cluster": "a", "filesystem": "fs", "user_group_ref": []string{"users"}}
			if _, err := build(Request{Action: action}, p); err != nil {
				t.Fatalf("valid %s ID %q rejected: %v", action, name, err)
			}
		}
		for _, name := range []string{"a_b", "a.b", "a b", "-a", "a-", "中文", strings.Repeat("a", 19)} {
			p := map[string]any{"name": name, "cluster": "a", "filesystem": "fs", "user_group_ref": []string{"users"}}
			if _, err := build(Request{Action: action}, p); err == nil {
				t.Fatalf("invalid %s ID %q accepted", action, name)
			}
		}
	}
	if _, err := build(Request{Action: "smb_share.create"}, map[string]any{"cluster": "bad_cluster", "name": "docs", "filesystem": "fs"}); err == nil {
		t.Fatal("invalid parent ID accepted")
	}
}
