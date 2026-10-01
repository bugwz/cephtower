package mutation

import (
	"cephtower/backend/internal/security"
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestNFSRGWUpdateExecution(t *testing.T) {
	service, _, id := newCephUserService(t)
	before := `[{"export_id":2,"cluster_id":"nfs-a","pseudo":"/share","path":"bucket","fsal":{"name":"RGW","user_id":"owner","access_key_id":"example-access","secret_access_key":"example-secret"}}]`
	runner := &directoryRenameExecutor{outputs: map[string]string{"nfs_export.update.pre_check": before, "nfs_export.update.post_check": before}}
	service.executor = runner
	p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "bucket", "fsal_type": "RGW", "rgw_user_id": "owner"}
	request := Request{ClusterID: id, Action: "nfs_export.update", ResourceKey: "nfs/export/" + base64.RawURLEncoding.EncodeToString([]byte("nfs-a\x002")), Parameters: p}
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 {
		t.Fatalf("unexpected command count: %d", len(runner.specs))
	}
	runner.outputs["nfs_export.update.post_check"] = `[]`
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("missing native RGW export accepted")
	}
}

func TestNFSRGWExport(t *testing.T) {
	p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "bucket", "fsal_type": "RGW", "rgw_user_id": "owner"}
	cmd, err := build(Request{Action: "nfs_export.create"}, p)
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(cmd.stdin, &record); err != nil {
		t.Fatal(err)
	}
	if !nfsExportAttributesMatch(record, p) {
		t.Fatal("RGW create payload mismatch")
	}
	fsal := record["fsal"].(map[string]any)
	fsal["secret_access_key"] = "test-secret"
	fsal["access_key_id"] = "test-access"
	data, err := nfsExportUpdateJSON(record, p)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	if record["fsal"].(map[string]any)["secret_access_key"] != "test-secret" {
		t.Fatal("native credentials lost during update")
	}
	redacted, err := security.RedactJSON(record)
	if err != nil {
		t.Fatal(err)
	}
	redactedFSAL := redacted.(map[string]any)["fsal"].(map[string]any)
	if redactedFSAL["secret_access_key"] == "test-secret" {
		t.Fatal("RGW secret exposed")
	}
	p["rgw_user_id"] = "other"
	if nfsExportAttributesMatch(record, p) {
		t.Fatal("wrong RGW user accepted")
	}
	p["fsal_type"] = "CEPH"
	p["filesystem"] = "fs"
	p["path"] = "/"
	if _, err := nfsExportUpdateJSON(record, p); err == nil {
		t.Fatal("FSAL switch accepted")
	}
	p["fsal_type"] = "RGW"
	delete(p, "rgw_user_id")
	if _, err := build(Request{Action: "nfs_export.create"}, p); err == nil {
		t.Fatal("missing RGW user accepted")
	}
}
