package mutation

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSMBClusterAuthSwitchRequiresSourcesBeforeApply(t *testing.T) {
	for _, mode := range []string{"user", "active-directory"} {
		t.Run(mode, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			old := "user"
			if mode == "user" {
				old = "active-directory"
			}
			before := `{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"` + old + `"}`
			runner := &directoryRenameExecutor{outputs: map[string]string{"smb_cluster.update.pre_check": before}}
			service.executor = runner
			request := Request{ClusterID: id, Action: "smb_cluster.update", ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": mode}}
			if _, err := service.Execute(context.Background(), request); err == nil {
				t.Fatal("mode-only switch accepted")
			}
			if len(runner.specs) != 1 || runner.specs[0].Mutating {
				t.Fatalf("incomplete switch reached mutation: %v", runner.specs)
			}
		})
	}
}

func TestSMBClusterUpdatePreservesSettings(t *testing.T) {
	service, _, id := newCephUserService(t)
	request := Request{ClusterID: id, Action: "smb_cluster.update", ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user"}}
	before := `{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","custom_dns":["192.0.2.1"],"placement":{"count":2},"user_group_settings":[{"source_type":"resource","ref":"users"}]}`
	data, err := smbClusterUpdateJSON([]byte(before), request)
	if err != nil || !smbClusterUpdateMatches([]byte(before), data, request) {
		t.Fatalf("settings not preserved: %s %v", data, err)
	}
	runner := &directoryRenameExecutor{outputs: map[string]string{"smb_cluster.update.pre_check": before, "smb_cluster.update": `{"success":true}`, "smb_cluster.update.post_check": before}}
	service.executor = runner
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 || runner.specs[1].Args[1] != "apply" || string(runner.specs[1].Stdin) != string(data) || runner.specs[0].Args[2] != "ceph.smb.cluster.a" {
		t.Fatal(runner.specs)
	}
	for _, bad := range []string{`null`, `[]`, before + `{}`, `{"resource_type":"ceph.smb.cluster","cluster_id":"other"}`} {
		if _, err := smbClusterUpdateJSON([]byte(bad), request); err == nil {
			t.Fatalf("invalid response accepted: %s", bad)
		}
	}
	runner.outputs["smb_cluster.update.post_check"] = `{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user"}`
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("lost settings accepted")
	}
	runner.outputs["smb_cluster.update"] = `{"success":false}`
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("failed apply accepted")
	}
}

func TestSMBClusterDNSUpdate(t *testing.T) {
	request := Request{ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user"}}
	before := []byte(`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","custom_dns":["192.0.2.1"]}`)
	for _, servers := range [][]string{{"192.0.2.2", "2001:db8::1"}, {}} {
		request.Parameters["custom_dns"] = servers
		data, err := smbClusterUpdateJSON(before, request)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]json.RawMessage
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		want, _ := json.Marshal(servers)
		if string(record["custom_dns"]) != string(want) || smbClusterUpdateMatches(data, before, request) {
			t.Fatalf("DNS update not verified: %s", data)
		}
	}
	for _, bad := range []any{nil, "192.0.2.1", []any{1}, []string{""}, []string{"not-an-ip"}, []string{"192.0.2.1/24"}} {
		request.Parameters["custom_dns"] = bad
		if _, err := smbClusterUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid DNS accepted: %v", bad)
		}
	}
}

func TestSMBClusterUserGroupReferences(t *testing.T) {
	request := Request{ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user", "user_group_ref": []string{"team", "ops"}}}
	before := []byte(`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"active-directory","domain_settings":{"realm":"EXAMPLE.COM"},"custom_dns":["192.0.2.1"]}`)
	data, err := smbClusterUpdateJSON(before, request)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	refs := record["user_group_settings"].([]any)
	if len(refs) != 2 || refs[0].(map[string]any)["ref"] != "team" || refs[0].(map[string]any)["source_type"] != "resource" || record["domain_settings"] != nil || record["custom_dns"] == nil {
		t.Fatal(record)
	}
	if !smbClusterUpdateMatches(data, data, request) || smbClusterUpdateMatches(data, before, request) {
		t.Fatal("reference readback failed")
	}
	for _, bad := range []any{nil, []string{}, []string{"team", "team"}, []string{"bad/id"}, []string{"abcdefghijklmnopqrs"}, "team", []any{1}} {
		request.Parameters["user_group_ref"] = bad
		if _, err := smbClusterUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid references accepted: %v", bad)
		}
	}
	request.Parameters["user_group_ref"] = []string{"team"}
	request.Parameters["auth_mode"] = "active-directory"
	if _, err := smbClusterUpdateJSON(before, request); err == nil {
		t.Fatal("local references accepted in AD mode")
	}
}

func TestSMBClusterDomainSettings(t *testing.T) {
	request := Request{ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "active-directory", "domain_realm": "EXAMPLE.COM", "domain_join_ref": []string{"join-a"}}}
	before := []byte(`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","user_group_settings":[{"source_type":"resource","ref":"users"}]}`)
	data, err := smbClusterUpdateJSON(before, request)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	domain := record["domain_settings"].(map[string]any)
	if domain["realm"] != "EXAMPLE.COM" || domain["join_sources"].([]any)[0].(map[string]any)["ref"] != "join-a" || record["user_group_settings"] != nil {
		t.Fatal(record)
	}
	if !smbClusterUpdateMatches(data, data, request) || smbClusterUpdateMatches(data, before, request) {
		t.Fatal("domain readback failed")
	}
	for _, bad := range []any{nil, []string{}, []string{"join-a", "join-a"}, []string{"bad/id"}, "join-a"} {
		request.Parameters["domain_join_ref"] = bad
		if _, err := smbClusterUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid join refs accepted: %v", bad)
		}
	}
	request.Parameters["domain_join_ref"] = []string{"join-a"}
	delete(request.Parameters, "domain_realm")
	if _, err := smbClusterUpdateJSON(before, request); err == nil {
		t.Fatal("missing realm accepted")
	}
	request.Parameters["domain_realm"] = "EXAMPLE.COM"
	request.Parameters["auth_mode"] = "user"
	if _, err := smbClusterUpdateJSON(before, request); err == nil {
		t.Fatal("domain in user mode accepted")
	}
}

func TestSMBClusterPublicAddressUpdate(t *testing.T) {
	service, _, id := newCephUserService(t)
	request := Request{ClusterID: id, Action: "smb_cluster.update", ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user"}}
	before := `{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","placement":{"count":2},"public_addrs":[{"address":"192.0.2.1/24","destination":["192.0.2.0/24","198.51.100.0/24"]}]}`
	unchanged, err := smbClusterUpdateJSON([]byte(before), request)
	if err != nil || !smbClusterUpdateMatches([]byte(before), unchanged, request) {
		t.Fatal("existing destinations not preserved", err)
	}
	for _, addresses := range [][]string{{"192.0.2.10/24%192.0.2.0/24", "2001:db8::10/64"}, {}} {
		request.Parameters["smb_public_addresses"] = addresses
		data, err := smbClusterUpdateJSON([]byte(before), request)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		items := record["public_addrs"].([]any)
		if len(items) != len(addresses) || record["placement"].(map[string]any)["count"] != float64(2) {
			t.Fatal(record)
		}
		if len(items) > 0 && (items[0].(map[string]any)["address"] != "192.0.2.10/24" || items[0].(map[string]any)["destination"] != "192.0.2.0/24" || items[1].(map[string]any)["address"] != "2001:db8::10/64") {
			t.Fatal(items)
		}
		runner := &directoryRenameExecutor{outputs: map[string]string{"smb_cluster.update.pre_check": before, "smb_cluster.update": `{"success":true}`, "smb_cluster.update.post_check": string(data)}}
		service.executor = runner
		if _, err := service.Execute(context.Background(), request); err != nil {
			t.Fatal(err)
		}
		if len(runner.specs) != 3 || string(runner.specs[1].Stdin) != string(data) {
			t.Fatal(runner.specs)
		}
		if smbClusterUpdateMatches(data, []byte(before), request) {
			t.Fatal("stale addresses accepted")
		}
	}
	request.Parameters["smb_public_addresses"] = []string{"invalid"}
	if _, err := smbClusterUpdateJSON([]byte(before), request); err == nil {
		t.Fatal("invalid address accepted")
	}
}

func TestSMBClusterPlacementLabel(t *testing.T) {
	service, _, id := newCephUserService(t)
	request := Request{ClusterID: id, Action: "smb_cluster.update", ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user", "smb_label": "smb"}}
	before := `{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","placement":{"count":2,"hosts":["node-a"],"host_pattern":"node*"}}`
	data, err := smbClusterUpdateJSON([]byte(before), request)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	placement := record["placement"].(map[string]any)
	if placement["label"] != "smb" || placement["count"] != float64(2) || placement["hosts"] != nil || placement["host_pattern"] != nil {
		t.Fatal(placement)
	}
	runner := &directoryRenameExecutor{outputs: map[string]string{"smb_cluster.update.pre_check": before, "smb_cluster.update": `{"success":true}`, "smb_cluster.update.post_check": string(data)}}
	service.executor = runner
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 || string(runner.specs[1].Stdin) != string(data) {
		t.Fatal(runner.specs)
	}
	placement["hosts"] = []string{"node-a"}
	stale, _ := json.Marshal(record)
	if smbClusterUpdateMatches(data, stale, request) {
		t.Fatal("stale hosts accepted")
	}
	request.Parameters["smb_hosts"] = []string{"node-a"}
	if _, err := smbClusterUpdateJSON([]byte(before), request); err == nil {
		t.Fatal("conflicting selectors accepted")
	}
	delete(request.Parameters, "smb_hosts")
	for _, bad := range []any{nil, "", "smb other", "smb:other"} {
		request.Parameters["smb_label"] = bad
		if _, err := smbClusterUpdateJSON([]byte(before), request); err == nil {
			t.Fatalf("invalid label accepted: %v", bad)
		}
	}
}

func TestSMBClusterPlacementHosts(t *testing.T) {
	request := Request{ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user", "smb_hosts": []string{"node-a", "node-b"}}}
	before := []byte(`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","placement":{"count":2,"label":"smb","host_pattern":"node*"}}`)
	data, err := smbClusterUpdateJSON(before, request)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	placement := record["placement"].(map[string]any)
	if placement["count"] != float64(2) || placement["label"] != nil || placement["host_pattern"] != nil || len(placement["hosts"].([]any)) != 2 {
		t.Fatal(placement)
	}
	if !smbClusterUpdateMatches(data, data, request) || smbClusterUpdateMatches(data, before, request) {
		t.Fatal("host changes not verified")
	}
	for _, bad := range []any{nil, []string{}, []string{"node-a", "node-a"}, []string{"label:smb"}, []string{"node*"}, []string{"123"}, "node-a"} {
		request.Parameters["smb_hosts"] = bad
		if _, err := smbClusterUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid hosts accepted: %v", bad)
		}
	}
}

func TestSMBClusterPlacementCount(t *testing.T) {
	request := Request{ResourceKey: "smb/cluster/a", Parameters: map[string]any{"auth_mode": "user", "count": json.Number("3")}}
	before := []byte(`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user","placement":{"count":2,"label":"smb","host_pattern":"node*"}}`)
	data, err := smbClusterUpdateJSON(before, request)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	placement := record["placement"].(map[string]any)
	if placement["count"] != float64(3) || placement["label"] != "smb" || placement["host_pattern"] != "node*" {
		t.Fatal(placement)
	}
	if smbClusterUpdateMatches(data, before, request) || !smbClusterUpdateMatches(data, data, request) {
		t.Fatal("count readback not verified")
	}
	for _, bad := range []any{nil, 0, -1, 1.5, true, "invalid"} {
		request.Parameters["count"] = bad
		if _, err := smbClusterUpdateJSON(before, request); err == nil {
			t.Fatalf("invalid count accepted: %v", bad)
		}
	}
}
