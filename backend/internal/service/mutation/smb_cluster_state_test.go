package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestSMBClusterLifecycleReadback(t *testing.T) {
	for _, action := range []string{"smb_cluster.create", "smb_cluster.delete"} {
		t.Run(action, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			request := Request{ClusterID: id, Action: action, ResourceKey: "smb/cluster/target", Parameters: map[string]any{"name": "target", "auth_mode": "user", "user_group_ref": []string{"users"}}}
			good, stale := `["other","target"]`, `["other"]`
			if action == "smb_cluster.delete" {
				good, stale = stale, good
				request.Parameters["name"] = "other"
			}
			runner := &directoryRenameExecutor{outputs: map[string]string{action + ".post_check": good}}
			service.executor = runner
			if _, err := service.Execute(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			if len(runner.specs) != 2 || runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, []string{"smb", "cluster", "ls", "--format", "json"}) {
				t.Fatalf("incorrect readback: %v", runner.specs)
			}
			for _, bad := range []string{stale, `null`, `{}`, `[1]`, `[""]`, `["other","other"]`, good + ` {}`} {
				runner.outputs[action+".post_check"] = bad
				_, err := service.Execute(context.Background(), request)
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Code != "post_check_failed" || !failure.Retryable {
					t.Fatalf("invalid readback %s accepted: %v", bad, err)
				}
			}
		})
	}
}
