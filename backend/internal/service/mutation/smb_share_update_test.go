package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSMBShareUpdatePreservesNativeSettings(t *testing.T) {
	service, _, id := newCephUserService(t)
	request := Request{ClusterID: id, Action: "smb_share.update", ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"cluster": "a", "filesystem": "fs", "path": "/new"}}
	before := `{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","name":"Documents","readonly":true,"browseable":false,"cephfs":{"volume":"fs","path":"/old","provider":"samba-vfs","subvolume":"sub","subvolumegroup":"group"}}`
	data, err := smbShareUpdateJSON([]byte(before), request)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	fs := payload["cephfs"].(map[string]any)
	if payload["readonly"] != true || payload["browseable"] != false || fs["path"] != "/new" || fs["subvolume"] != "sub" || fs["subvolumegroup"] != "group" {
		t.Fatal(payload)
	}
	runner := &directoryRenameExecutor{outputs: map[string]string{"smb_share.update.pre_check": before, "smb_share.update.post_check": string(data), "smb_share.update": `{"success":true}`}}
	service.executor = runner
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 || string(runner.specs[1].Stdin) != string(data) || runner.specs[1].Args[1] != "apply" {
		t.Fatal(runner.specs)
	}
	runner.outputs["smb_share.update.post_check"] = before
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("stale readback accepted")
	}
	for _, bad := range []string{`null`, `[]`, before + ` {}`, `{"resource_type":"ceph.smb.share","cluster_id":"other","share_id":"docs","cephfs":{}}`} {
		if _, err := smbShareUpdateJSON([]byte(bad), request); err == nil {
			t.Fatal("invalid identity accepted")
		}
	}
	delete(request.Parameters, "path")
	runner.outputs["smb_share.update"] = `{"success":false}`
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("native apply failure accepted even though unchanged readback matches")
	}
	preserved, err := smbShareUpdateJSON([]byte(before), request)
	if err != nil || !smbShareUpdateMatches(preserved, []byte(before), request) {
		t.Fatal("omitted path changed")
	}
}
