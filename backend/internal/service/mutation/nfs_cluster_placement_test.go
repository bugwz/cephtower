package mutation

import (
	"reflect"
	"strings"
	"testing"
)

func TestNFSClusterPlacement(t *testing.T) {
	for _, placement := range []string{"2 host-a host-b", "label:nfs", "3", "host*", " host-a "} {
		spec, err := build(Request{Action: "nfs_cluster.create"}, map[string]any{"name": "nfs-a", "nfs_placement": placement})
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"nfs", "cluster", "create", "nfs-a", "--placement=" + strings.TrimSpace(placement)}
		if !reflect.DeepEqual(spec.args, want) {
			t.Fatalf("args=%v", spec.args)
		}
	}
	for _, placement := range []any{nil, 42, "", "  ", "host\n--bad", "host\x00", strings.Repeat("a", 1025)} {
		if _, err := build(Request{Action: "nfs_cluster.create"}, map[string]any{"name": "nfs-a", "nfs_placement": placement}); err == nil {
			t.Fatalf("accepted %v", placement)
		}
	}
	spec, err := build(Request{Action: "nfs_cluster.create"}, map[string]any{"name": "nfs-a"})
	if err != nil || !reflect.DeepEqual(spec.args, []string{"nfs", "cluster", "create", "nfs-a"}) {
		t.Fatal("omitted placement changed native default")
	}
}
