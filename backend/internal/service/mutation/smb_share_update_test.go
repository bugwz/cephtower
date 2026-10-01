package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSMBShareScopeReplacement(t *testing.T) {
	service, _, id := newCephUserService(t)
	request := Request{ClusterID: id, Action: "smb_share.update", ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"cluster": "a", "filesystem": "other", "path": "/target"}}
	before := `{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","readonly":true,"cephfs":{"volume":"fs","path":"/","subvolume":"old","subvolumegroup":"oldgroup","provider":"samba-vfs"}}`
	for _, scope := range []string{"group/new", "new", ""} {
		request.Parameters["subvolume"] = scope
		data, err := smbShareUpdateJSON([]byte(before), request)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(data, &record); err != nil {
			t.Fatal(err)
		}
		fs := record["cephfs"].(map[string]any)
		if fs["volume"] != "other" || fs["path"] != "/target" || fs["provider"] != "samba-vfs" || record["readonly"] != true {
			t.Fatal(record)
		}
		if scope == "group/new" && (fs["subvolume"] != "new" || fs["subvolumegroup"] != "group") {
			t.Fatal(fs)
		}
		if scope == "new" && (fs["subvolume"] != "new" || fs["subvolumegroup"] != nil) {
			t.Fatal(fs)
		}
		if scope == "" && (fs["subvolume"] != nil || fs["subvolumegroup"] != nil) {
			t.Fatal(fs)
		}
		runner := &directoryRenameExecutor{outputs: map[string]string{"smb_share.update.pre_check": before, "smb_share.update": `{"success":true}`, "smb_share.update.post_check": string(data)}}
		service.executor = runner
		if _, err := service.Execute(context.Background(), request); err != nil {
			t.Fatal(err)
		}
		if len(runner.specs) != 3 || string(runner.specs[1].Stdin) != string(data) {
			t.Fatal(runner.specs)
		}
		fs["subvolumegroup"] = "oldgroup"
		stale, _ := json.Marshal(record)
		if smbShareUpdateMatches(data, stale, request) {
			t.Fatal("stale group accepted")
		}
	}
	for _, bad := range []any{nil, false, "a/b/c", "/sub", "group/", "..", "a\nb"} {
		request.Parameters["subvolume"] = bad
		if _, err := smbShareUpdateJSON([]byte(before), request); err == nil {
			t.Fatalf("invalid scope accepted: %v", bad)
		}
	}
	request.Parameters["subvolume"] = ""
	delete(request.Parameters, "path")
	if _, err := smbShareUpdateJSON([]byte(before), request); err == nil {
		t.Fatal("scope replacement without explicit path accepted")
	}
}

func TestSMBShareFilesystemScope(t *testing.T) {
	for _, scope := range []string{`"subvolume":"sub"`, `"subvolumegroup":"group"`, `"subvolume":"sub","subvolumegroup":"group"`} {
		service, _, id := newCephUserService(t)
		request := Request{ClusterID: id, Action: "smb_share.update", ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"cluster": "a", "filesystem": "other", "path": "/new"}}
		before := `{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","cephfs":{"volume":"fs","path":"/",` + scope + `}}`
		runner := &directoryRenameExecutor{outputs: map[string]string{"smb_share.update.pre_check": before}}
		service.executor = runner
		if _, err := service.Execute(context.Background(), request); err == nil {
			t.Fatal("cross-filesystem subvolume retained")
		}
		if len(runner.specs) != 1 || runner.specs[0].Mutating {
			t.Fatal("unsafe scope reached mutation", runner.specs)
		}
		request.Parameters["filesystem"] = "fs"
		if _, err := smbShareUpdateJSON([]byte(before), request); err != nil {
			t.Fatal("same filesystem rejected", err)
		}
	}
	request := Request{ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"filesystem": "other"}}
	before := []byte(`{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","cephfs":{"volume":"fs","path":"/","subvolume":"","subvolumegroup":""}}`)
	data, err := smbShareUpdateJSON(before, request)
	if err != nil {
		t.Fatal("unscoped filesystem change rejected", err)
	}
	var record map[string]any
	if err := json.Unmarshal(data, &record); err != nil || record["cephfs"].(map[string]any)["volume"] != "other" {
		t.Fatal("filesystem not changed", err)
	}
}

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
	for _, flag := range []bool{false, true} {
		request.Parameters["readonly"] = flag
		request.Parameters["browseable"] = flag
		changed, err := smbShareUpdateJSON([]byte(before), request)
		if err != nil {
			t.Fatal(err)
		}
		var record map[string]any
		if err := json.Unmarshal(changed, &record); err != nil {
			t.Fatal(err)
		}
		if record["readonly"] != flag || record["browseable"] != flag {
			t.Fatal("boolean lost")
		}
		if !smbShareUpdateMatches(changed, changed, request) {
			t.Fatal("readback rejected")
		}
	}
	request.Parameters["readonly"] = "false"
	if _, err := smbShareUpdateJSON([]byte(before), request); err == nil {
		t.Fatal("non-boolean accepted")
	}
	delete(request.Parameters, "readonly")
	delete(request.Parameters, "browseable")
	runner.outputs["smb_share.update"] = `{"success":false}`
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("native apply failure accepted even though unchanged readback matches")
	}
	preserved, err := smbShareUpdateJSON([]byte(before), request)
	if err != nil || !smbShareUpdateMatches(preserved, []byte(before), request) {
		t.Fatal("omitted path changed")
	}
}
