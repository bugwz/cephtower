package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWUserPresenceRequiresCompleteNativeList(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"keys":[],"truncated":true}`, `{"keys":[],"truncated":false}`, `[null]`, `["u","u"]`, `[""]`, `[1]`, `["u"`} {
		for _, present := range []bool{true, false} {
			if rgwUserPresence([]byte(raw), "u", present) {
				t.Fatalf("accepted invalid list %s", raw)
			}
		}
	}
	if !rgwUserPresence([]byte(`[]`), "u", false) || !rgwUserPresence([]byte(`["tenant$ns$u","other"]`), "tenant$ns$u", true) || rgwUserPresence([]byte(`["u"]`), "tenant$ns$u", true) {
		t.Fatal("presence did not preserve exact UID")
	}
}

func TestRGWUserDeletionVerification(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, scenario := range []string{"success", "absent", "pre-error", "invalid-pre", "write-error", "post-error", "retained", "invalid-post", "paged-post"} {
		t.Run(scenario, func(t *testing.T) {
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.delete.pre_check": `["tenant$ns$user","other"]`, "rgw_user.delete.post_check": `["other"]`}}
			count := 3
			switch scenario {
			case "absent":
				runner.outputs["rgw_user.delete.pre_check"] = `["user"]`
				count = 1
			case "pre-error":
				runner.failID = "rgw_user.delete.pre_check"
				count = 1
			case "invalid-pre":
				runner.outputs["rgw_user.delete.pre_check"] = `null`
				count = 1
			case "write-error":
				runner.failID = "rgw_user.delete"
				count = 2
			case "post-error":
				runner.failID = "rgw_user.delete.post_check"
			case "retained":
				runner.outputs["rgw_user.delete.post_check"] = `["tenant$ns$user"]`
			case "invalid-post":
				runner.outputs["rgw_user.delete.post_check"] = `{}`
			case "paged-post":
				runner.outputs["rgw_user.delete.post_check"] = `{"keys":[],"truncated":true}`
			}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.delete", ResourceKey: "rgw/user/tenant$ns$user"})
			if (err == nil) != (scenario == "success") || len(runner.specs) != count {
				t.Fatalf("%v %+v", err, runner.specs)
			}
			if err != nil && count > 1 {
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Retryable || (count == 3 && failure.Code != "post_check_failed") {
					t.Fatalf("unsafe result: %v", err)
				}
			}
			for i, spec := range runner.specs {
				want := []string{"user", "list", "--format", "json"}
				if i == 1 {
					want = []string{"user", "rm", "--uid", "tenant$ns$user", "--format", "json"}
				}
				if !reflect.DeepEqual(spec.Args, want) || spec.Mutating != (i == 1) {
					t.Fatalf("unexpected command: %+v", spec)
				}
			}
		})
	}
}
