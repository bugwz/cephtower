package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"testing"
)

func TestNFSClusterStateReadback(t *testing.T) {
	for _, action := range []string{"nfs_cluster.create", "nfs_cluster.delete"} {
		request := Request{Action: action, ResourceKey: "nfs/cluster/target", Parameters: map[string]any{"name": "target"}}
		for _, tc := range []struct {
			data           string
			create, remove bool
		}{
			{`["target"]`, true, false}, {`[]`, false, true}, {`["other"]`, false, true},
			{`null`, false, false}, {`{}`, false, false}, {`[null]`, false, false},
			{`["target","target"]`, false, false}, {`[""]`, false, false},
			{`["target"] []`, false, false}, {`["target",3]`, false, false},
		} {
			want := tc.create
			if action == "nfs_cluster.delete" {
				want = tc.remove
			}
			if got := nfsClusterStateMatches(request, []byte(tc.data)); got != want {
				t.Fatalf("%s %s got %v", action, tc.data, got)
			}
		}
	}
}

func TestNFSClusterExecuteRejectsUnverifiedState(t *testing.T) {
	service, _, id := newCephUserService(t)
	for _, action := range []string{"nfs_cluster.create", "nfs_cluster.delete"} {
		request := Request{ClusterID: id, Action: action, ResourceKey: "nfs/cluster/target", Parameters: map[string]any{"name": "target"}}
		good, bad := `["target"]`, `[]`
		if action == "nfs_cluster.delete" {
			good, bad = bad, good
		}
		runner := &directoryRenameExecutor{outputs: map[string]string{action + ".post_check": bad, action + ".pre_check": `[]`}}
		service.executor = runner
		_, err := service.Execute(context.Background(), request)
		var actionErr *cephdomain.ActionError
		if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" {
			t.Fatalf("unverified state accepted: %v", err)
		}
		runner.outputs[action+".post_check"] = good
		if _, err := service.Execute(context.Background(), request); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNFSClusterCreatePrecheck(t *testing.T) {
	service, _, id := newCephUserService(t)
	for _, data := range []string{`["target"]`, `null`, `{}`, `[] []`, `["other","other"]`} {
		runner := &directoryRenameExecutor{outputs: map[string]string{"nfs_cluster.create.pre_check": data}}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "nfs_cluster.create", ResourceKey: "nfs/cluster", Parameters: map[string]any{"name": "target", "nfs_port": 2050}})
		if err == nil || len(runner.specs) != 1 || runner.specs[0].Mutating {
			t.Fatalf("unsafe creation reached mutation for %s", data)
		}
	}
}
