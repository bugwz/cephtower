package mutation

import (
	"context"
	"encoding/json"
	"testing"
)

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
