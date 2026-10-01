package mutation

import (
	"reflect"
	"slices"
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

func TestSMBClusterCreateDNS(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"users"}, "custom_dns": []string{"192.0.2.1", "2001:db8::1"}}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || !slices.Contains(spec.args, "--custom-dns=192.0.2.1") || !slices.Contains(spec.args, "--custom-dns=2001:db8::1") {
		t.Fatalf("DNS not passed: %v %v", spec.args, err)
	}
	for _, bad := range []any{nil, "192.0.2.1", []string{"bad"}, []string{"192.0.2.1/24"}, []any{1}} {
		p["custom_dns"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid DNS accepted: %v", bad)
		}
	}
	p["custom_dns"] = []string{}
	if _, err := build(Request{Action: "smb_cluster.create"}, p); err != nil {
		t.Fatal(err)
	}
}

func TestSMBClusterCreateCount(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"users"}, "count": float64(3)}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || !slices.Contains(spec.args, "--placement=count:3") {
		t.Fatalf("count not passed: %v %v", spec.args, err)
	}
	for _, bad := range []any{nil, 0, -1, 1.5, true, "bad"} {
		p["count"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid count accepted: %v", bad)
		}
	}
	delete(p, "count")
	spec, err = build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || len(spec.args) != 6 {
		t.Fatalf("default placement changed: %v %v", spec.args, err)
	}
}

func TestSMBClusterCreateHosts(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"users"}, "count": float64(2), "smb_hosts": []string{"node-a", "node-b"}}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || !slices.Contains(spec.args, "--placement=count:2 node-a node-b") {
		t.Fatalf("placement not combined: %v %v", spec.args, err)
	}
	delete(p, "count")
	spec, err = build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || !slices.Contains(spec.args, "--placement=node-a node-b") {
		t.Fatalf("host placement failed: %v %v", spec.args, err)
	}
	for _, bad := range []any{nil, []string{}, "node-a", []string{"node-a", "node-a"}, []string{"label:smb"}, []string{"node-a count:5"}, []string{"*"}, []string{"123"}} {
		p["smb_hosts"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid hosts accepted: %v", bad)
		}
	}
}

func TestSMBClusterCreateDomainReferences(t *testing.T) {
	p := map[string]any{"name": "smb-a", "auth_mode": "active-directory", "domain_realm": "EXAMPLE.COM", "domain_join_ref": []string{"join-a", "join-b"}}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	want := []string{"smb", "cluster", "create", "smb-a", "active-directory", "--domain-realm=EXAMPLE.COM", "--domain-join-ref=join-a", "--domain-join-ref=join-b"}
	if err != nil || !reflect.DeepEqual(spec.args, want) {
		t.Fatalf("wrong AD command: %v %v", spec.args, err)
	}
	for _, bad := range []any{nil, []string{}, "join-a", []string{"join-a", "join-a"}, []string{"bad/id"}} {
		p["domain_join_ref"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid join references accepted: %v", bad)
		}
	}
	p["domain_join_ref"] = []string{"join-a"}
	for _, bad := range []any{nil, "", "\n", "example\x00.com"} {
		p["domain_realm"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid realm accepted: %v", bad)
		}
	}
	p["domain_realm"] = "EXAMPLE.COM"
	p["user_group_ref"] = []string{"users"}
	if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
		t.Fatal("mixed authentication sources accepted")
	}
}
