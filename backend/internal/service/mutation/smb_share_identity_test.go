package mutation

import (
	"encoding/base64"
	"testing"
)

func TestSMBShareUpdateIdentity(t *testing.T) {
	request := Request{Action: "smb_share.update", ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("cluster-a\x00share-a"))}
	parameters := map[string]any{"cluster": "cluster-a", "filesystem": "fs", "name": "ignored"}
	spec, err := build(request, parameters)
	if err != nil || spec.check[2] != "ceph.smb.share.cluster-a.share-a" {
		t.Fatalf("identity not preserved: %v %v", spec.args, err)
	}
	parameters["cluster"] = "cluster-b"
	if _, err := build(request, parameters); err == nil {
		t.Fatal("cross-cluster update accepted")
	}
}
