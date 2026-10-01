package mutation

import (
	"reflect"
	"slices"
	"testing"
)

func TestSMBClusterPublicAddresses(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"users"}, "smb_public_addresses": []string{"192.0.2.10/24%192.0.2.0/24", "2001:db8::10/64"}}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || !slices.Contains(spec.args, "--public-addrs=192.0.2.10/24%192.0.2.0/24") || !slices.Contains(spec.args, "--public-addrs=2001:db8::10/64") {
		t.Fatalf("wrong public addresses: %v %v", spec.args, err)
	}
	for _, bad := range []any{nil, "192.0.2.1/24", []string{"192.0.2.1"}, []string{"192.0.2.1/33"}, []string{"192.0.2.1/24%192.0.2.2/24"}, []string{"192.0.2.1/24%"}, []string{"192.0.2.1/24%x%y"}} {
		p["smb_public_addresses"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid address accepted: %v", bad)
		}
	}
	p["smb_public_addresses"] = []string{}
	if _, err := build(Request{Action: "smb_cluster.create"}, p); err != nil {
		t.Fatal(err)
	}
}

func TestSMBClusterLabelPlacement(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"users"}, "smb_label": "smb", "count": float64(2)}
	spec, err := build(Request{Action: "smb_cluster.create"}, p)
	if err != nil || !slices.Contains(spec.args, "--placement=label:smb count:2") {
		t.Fatalf("wrong label placement: %v %v", spec.args, err)
	}
	p["smb_hosts"] = []string{"node-a"}
	if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
		t.Fatal("conflicting placement accepted")
	}
	delete(p, "smb_hosts")
	for _, bad := range []any{nil, "", "smb count:5", "smb,other", "smb\nnode-a"} {
		p["smb_label"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid label accepted: %v", bad)
		}
	}
}

func TestSMBClusterClusteringMode(t *testing.T) {
	p := map[string]any{"name": "smb-a", "user_group_ref": []string{"users"}}
	request := Request{ResourceKey: "smb/cluster/smb-a", Parameters: p}
	p["auth_mode"] = "user"
	before := []byte(`{"resource_type":"ceph.smb.cluster","cluster_id":"smb-a","auth_mode":"user","clustering":"never"}`)
	for _, mode := range []string{"default", "always", "never"} {
		p["clustering"] = mode
		spec, err := build(Request{Action: "smb_cluster.create"}, p)
		if err != nil || !slices.Contains(spec.args, "--clustering="+mode) {
			t.Fatalf("wrong clustering command: %v %v", spec.args, err)
		}
		data, err := smbClusterUpdateJSON(before, request)
		if err != nil || !smbClusterUpdateMatches(data, data, request) {
			t.Fatalf("invalid update: %s %v", data, err)
		}
		if mode != "never" && smbClusterUpdateMatches(data, before, request) {
			t.Fatal("stale mode accepted")
		}
	}
	for _, bad := range []any{nil, true, "", "invalid"} {
		p["clustering"] = bad
		if _, err := build(Request{Action: "smb_cluster.create"}, p); err == nil {
			t.Fatalf("invalid mode accepted: %v", bad)
		}
		if _, err := smbClusterUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid update accepted: %v", bad)
		}
	}
}

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
