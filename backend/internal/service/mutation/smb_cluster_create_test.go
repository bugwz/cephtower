package mutation

import (
	"reflect"
	"testing"
)

func TestSMBClusterCreateUserReferences(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"team", "ops"}}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	want := []string{"smb", "cluster", "create", "smb-a", "user", "--user-group-ref=team", "--user-group-ref=ops"}
	if err != nil || !reflect.DeepEqual(spec.args, want) {
		t.Fatalf("wrong native command: %v %v", spec.args, err)
	}
	for _, bad := range []any{nil, []string{}, "team", []string{"team", "team"}, []string{"bad/id"}, []any{1}} {
		p["user_group_ref"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid references accepted: %v", bad)
		}
	}
	delete(p, "user_group_ref")
	if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
		t.Fatal("missing local authentication accepted")
	}
	p["user_group_ref"] = []string{"team"}
	p["auth_mode"] = "active-directory"
	if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
		t.Fatal("local references in AD mode accepted")
	}
}
