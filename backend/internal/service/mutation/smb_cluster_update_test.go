package mutation

import (
	"context"
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
