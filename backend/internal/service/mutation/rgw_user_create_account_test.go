package mutation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWUserCreateInAccount(t *testing.T) {
	const account = "RGW12345678901234567"
	service, _, clusterID := newCephUserService(t)
	for _, root := range []bool{false, true} {
		kind := "rgw"
		if root {
			kind = "root"
		}
		for _, uid := range []string{"user", "tenant$user", "tenant$namespace$user", "$namespace$user"} {
			params := map[string]any{"uid": uid, "display_name": "valid-name", "account_id": account, "account_root": root}
			for _, tc := range []struct {
				response, fail string
				valid          bool
			}{
				{`{"full_user_id":"` + uid + `","display_name":"valid-name","account_id":"` + account + `","type":"` + kind + `","keys":[],"swift_keys":[]}`, "", true},
				{`{}`, "", false}, {`{}`, "rgw_user.create", false}, {`{}`, "rgw_user.create.post_check", false},
			} {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.create.pre_check": `[]`, "rgw_user.create.post_check": tc.response}, failID: tc.fail}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.create", ResourceKey: "rgw/user/" + uid, Parameters: params})
				if (err == nil) != tc.valid {
					t.Fatalf("root=%v uid=%s: %v", root, uid, err)
				}
				if err != nil {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Retryable {
						t.Fatalf("unsafe retry: %v", err)
					}
				}
				want := []string{"user", "create", "--uid", uid, "--account-id=" + account, "--account-root=" + strconv.FormatBool(root), "--display-name", "valid-name", "--generate-key=false", "--format", "json"}
				if !reflect.DeepEqual(runner.specs[1].Args, want) || !runner.specs[1].Mutating {
					t.Fatalf("unexpected write: %+v", runner.specs[1])
				}
				if tc.fail == "rgw_user.create" {
					if len(runner.specs) != 2 {
						t.Fatal("read after failed create")
					}
				} else if len(runner.specs) != 3 || runner.specs[2].Mutating || !reflect.DeepEqual(runner.specs[2].Args, []string{"user", "info", "--uid", uid, "--format", "json"}) {
					t.Fatal("wrong readback")
				}
			}
		}
	}
	for _, params := range []map[string]any{
		{"uid": "test", "display_name": "valid", "account_id": account},
		{"uid": "test", "display_name": "invalid name", "account_id": account, "account_root": false},
		{"uid": "test", "display_name": "valid", "account_root": true},
		{"uid": "test", "display_name": "valid", "account_id": account + " ", "account_root": true},
		{"uid": "test", "display_name": "valid", "account_id": account, "account_root": "false"},
	} {
		if _, err := build(Request{Action: "rgw_user.create"}, params); err == nil {
			t.Fatalf("accepted invalid account creation: %v", params)
		}
	}
	for _, uid := range []any{nil, false, "", " user", "user ", "-user", "user/name", "user\n", "user;name"} {
		if _, err := build(Request{Action: "rgw_user.create"}, map[string]any{"uid": uid, "display_name": "valid"}); err == nil {
			t.Fatalf("accepted invalid UID: %v", uid)
		}
	}
}
