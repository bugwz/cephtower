package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWCapsReplacement(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, permission := range []string{"read", "write", "read,write", "*"} {
		p := map[string]any{"uid": "tenant$ns$user", "action": "replace", "type": "users", "permission": permission}
		for _, old := range []string{"<none>", "read", "write", "*"} {
			for _, scenario := range []string{"success", "absent", "pre-error", "remove-error", "add-error", "post-error", "wrong-state", "other-changed"} {
				t.Run(permission+"/"+old+"/"+scenario, func(t *testing.T) {
					final := permission
					if final == "read,write" {
						final = "*"
					}
					runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.caps.pre_check": capsInfo(map[string]string{"users": old, "usage": "write"}), "rgw_user.caps.post_check": capsInfo(map[string]string{"users": final, "usage": "write"})}}
					count := 4
					switch scenario {
					case "absent":
						runner.outputs["rgw_user.caps.pre_check"] = capsInfo(map[string]string{"usage": "write"})
						count = 1
					case "pre-error":
						runner.failID = "rgw_user.caps.pre_check"
						count = 1
					case "remove-error":
						runner.failID = "rgw_user.caps"
						count = 2
					case "add-error":
						runner.failID = "rgw_user.caps.step2"
						count = 3
					case "post-error":
						runner.failID = "rgw_user.caps.post_check"
					case "wrong-state":
						runner.outputs["rgw_user.caps.post_check"] = capsInfo(map[string]string{"usage": "write"})
					case "other-changed":
						runner.outputs["rgw_user.caps.post_check"] = capsInfo(map[string]string{"users": final, "usage": "read"})
					}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.caps", ResourceKey: "rgw/user/tenant$ns$user", Parameters: p})
					if (err == nil) != (scenario == "success") || len(runner.specs) != count {
						t.Fatalf("%v %+v", err, runner.specs)
					}
					if err != nil && count > 1 {
						var failure *cephdomain.ActionError
						if !errors.As(err, &failure) || failure.Retryable {
							t.Fatalf("unsafe retry: %v", err)
						}
					}
					for i, spec := range runner.specs {
						want := []string{"user", "info", "--uid", "tenant$ns$user", "--format", "json"}
						if i == 1 {
							want = []string{"caps", "rm", "--uid", "tenant$ns$user", "--caps", "users=*", "--format", "json"}
						}
						if i == 2 {
							want = []string{"caps", "add", "--uid", "tenant$ns$user", "--caps", "users=" + permission, "--format", "json"}
						}
						if !reflect.DeepEqual(spec.Args, want) || spec.Mutating != (i == 1 || i == 2) {
							t.Fatalf("bad step: %+v", spec)
						}
					}
				})
			}
		}
	}
}
