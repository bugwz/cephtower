package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
)

func TestNFSScopedBucketUpdate(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &directoryRenameExecutor{outputs: map[string]string{
		"nfs_export.update.bucket_owner": `{"bucket":"same","tenant":"new","owner":"new$user"}`,
		"nfs_export.update.pre_check":    `[{"export_id":2,"cluster_id":"nfs-a","pseudo":"/share","path":"same","fsal":{"name":"RGW","user_id":"old$user","access_key_id":"old-access","secret_access_key":"old-secret"}}]`,
		"nfs_export.update.post_check":   `[{"export_id":2,"cluster_id":"nfs-a","pseudo":"/share","path":"same","fsal":{"name":"RGW","user_id":"new$user"}}]`,
	}}
	service.executor = runner
	request := Request{ClusterID: id, Action: "nfs_export.update", ResourceKey: "nfs/export/" + base64.RawURLEncoding.EncodeToString([]byte("nfs-a\x002")), Parameters: map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "same", "fsal_type": "RGW", "rgw_bucket_tenant": "new"}}
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(runner.specs[2].Stdin, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["fsal"].(map[string]any)["user_id"] != "new$user" {
		t.Fatal("tenant switch lost")
	}
	runner.outputs["nfs_export.update.post_check"] = runner.outputs["nfs_export.update.pre_check"]
	if _, err := service.Execute(context.Background(), request); err == nil {
		t.Fatal("wrong tenant readback accepted")
	}
}

func TestNFSScopedBucketOwner(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &directoryRenameExecutor{outputs: map[string]string{
		"nfs_export.create.bucket_owner": `{"bucket":"same","tenant":"tenant","owner":"tenant$user"}`,
		"nfs_export.create.pre_check":    `[]`,
		"nfs_export.create.post_check":   `[{"export_id":1,"cluster_id":"nfs-a","pseudo":"/share","path":"same","fsal":{"name":"RGW","user_id":"tenant$user"}}]`,
	}}
	service.executor = runner
	p := map[string]any{"cluster": "nfs-a", "pseudo": "/share", "path": "same", "fsal_type": "RGW", "rgw_bucket_tenant": "tenant"}
	request := Request{ClusterID: id, Action: "nfs_export.create", ResourceKey: "nfs/export", Parameters: p}
	if _, err := service.Execute(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runner.specs[0].Args, []string{"bucket", "stats", "--bucket", "same", "--format", "json", "--tenant", "tenant"}) {
		t.Fatal("missing scoped bucket query")
	}
	var payload map[string]any
	if err := json.Unmarshal(runner.specs[2].Stdin, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["fsal"].(map[string]any)["user_id"] != "tenant$user" {
		t.Fatal("owner not applied")
	}
	if _, exists := p["rgw_user_id"]; exists {
		t.Fatal("request was mutated")
	}
	for _, output := range []string{`{"bucket":"same","tenant":"","owner":"user"}`, `{"bucket":"other","tenant":"tenant","owner":"tenant$user"}`, `{"bucket":"same","tenant":"tenant","owner":"other$user"}`, `null`, `{} extra`} {
		runner.outputs["nfs_export.create.bucket_owner"] = output
		runner.specs = nil
		if _, err := service.Execute(context.Background(), request); err == nil || len(runner.specs) != 1 {
			t.Fatal("invalid bucket owner reached mutation")
		}
	}
}
