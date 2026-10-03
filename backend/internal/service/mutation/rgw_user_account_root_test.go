package mutation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserAccountRootExecution(t *testing.T) {
	const account = "RGW12345678901234567"
	service, _, clusterID := newCephUserService(t)
	info := func(kind string) string {
		return `{"full_user_id":"tenant$user","account_id":"` + account + `","type":"` + kind + `"}`
	}
	for _, root := range []bool{false, true} {
		kind := "rgw"
		if root {
			kind = "root"
		}
		for _, tc := range []struct {
			name, before, after, fail string
			success                   bool
			calls                     int
		}{
			{"valid", info("rgw"), info(kind), "", true, 3},
			{"existing root", info("root"), info(kind), "", true, 3},
			{"unknown membership", `{}`, info(kind), "", false, 1},
			{"wrong account", `{"full_user_id":"tenant$user","account_id":"RGW00000000000000000","type":"rgw"}`, info(kind), "", false, 1},
			{"unsupported type", info("ldap"), info(kind), "", false, 1},
			{"missing readback", info("rgw"), `{}`, "", false, 3},
			{"pre read failure", info("rgw"), info(kind), "rgw_user.update.pre_check", false, 1},
			{"write failure", info("rgw"), info(kind), "rgw_user.update", false, 2},
			{"post read failure", info("rgw"), info(kind), "rgw_user.update.post_check", false, 3},
		} {
			t.Run(strconv.FormatBool(root)+"/"+tc.name, func(t *testing.T) {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.update.pre_check": tc.before, "rgw_user.update.post_check": tc.after}, failID: tc.fail}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user", Parameters: map[string]any{"account_root": root, "expected_account_id": account}})
				if (err == nil) != tc.success {
					t.Fatalf("unexpected result: %v", err)
				}
				if !tc.success && tc.calls == 3 {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
						t.Fatalf("unexpected post-check error: %v", err)
					}
				}
				if len(runner.specs) != tc.calls {
					t.Fatalf("unexpected calls: %v", runner.specs)
				}
				for i, spec := range runner.specs {
					want := []string{"user", "info", "--uid", "tenant$user", "--format", "json"}
					if i == 1 {
						want = []string{"user", "modify", "--uid", "tenant$user", "--account-root=" + strconv.FormatBool(root), "--format", "json"}
					}
					if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i == 1) || !reflect.DeepEqual(spec.Args, want) {
						t.Fatalf("wrong step %d: %+v", i, spec)
					}
				}
			})
		}
		wrong := !root
		if rgwUserAccountRootMatches([]byte(info(kind)), "tenant$user", account, &wrong) {
			t.Fatal("accepted wrong root state")
		}
		if rgwUserAccountRootMatches([]byte(info(kind)), "other$user", account, &root) {
			t.Fatal("accepted wrong identity")
		}
	}
	for _, params := range []map[string]any{
		{"account_root": nil, "expected_account_id": account}, {"account_root": "true", "expected_account_id": account},
		{"account_root": true}, {"account_root": false, "expected_account_id": ""}, {"account_root": true, "expected_account_id": " " + account}, {"expected_account_id": account},
	} {
		if _, err := build(Request{Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user"}, params); err == nil {
			t.Fatalf("accepted invalid params: %v", params)
		}
	}
}
