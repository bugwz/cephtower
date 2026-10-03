package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserAccountMigration(t *testing.T) {
	const id = "RGW12345678901234567"
	const user = `{"full_user_id":"tenant$user","account_id":"","tenant":"tenant","display_name":"valid-name","type":"rgw"}`
	const account = `{"id":"RGW12345678901234567","tenant":"tenant"}`
	const after = `{"full_user_id":"tenant$user","account_id":"RGW12345678901234567","type":"rgw"}`
	service, _, clusterID := newCephUserService(t)
	for _, tc := range []struct {
		name, user, account, after, fail string
		calls                            int
		success                          bool
	}{
		{"valid", user, account, after, "", 4, true},
		{"assigned", strings.Replace(user, `"account_id":""`, `"account_id":"`+id+`"`, 1), account, after, "", 1, false},
		{"wrong user", strings.Replace(user, "tenant$user", "other$user", 1), account, after, "", 1, false},
		{"invalid name", strings.Replace(user, "valid-name", "invalid name", 1), account, after, "", 1, false},
		{"missing user fields", `{}`, account, after, "", 1, false},
		{"wrong tenant", user, strings.Replace(account, "tenant\"}", "other\"}", 1), after, "", 2, false},
		{"wrong account", user, strings.Replace(account, id, "RGW00000000000000000", 1), after, "", 2, false},
		{"missing account", user, `{}`, after, "", 2, false},
		{"user read failed", user, account, after, "rgw_user.update.migration_user", 1, false},
		{"account read failed", user, account, after, "rgw_user.update.migration_account", 2, false},
		{"partial write failure", user, account, after, "rgw_user.update", 3, false},
		{"readback failed", user, account, after, "rgw_user.update.post_check", 4, false},
		{"membership unchanged", user, account, user, "", 4, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.update.migration_user": tc.user, "rgw_user.update.migration_account": tc.account, "rgw_user.update.post_check": tc.after}, failID: tc.fail}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user", Parameters: map[string]any{"target_account_id": id, "migration_confirm_uid": "tenant$user"}})
			if (err == nil) != tc.success || len(runner.specs) != tc.calls {
				t.Fatalf("err=%v calls=%v", err, runner.specs)
			}
			if !tc.success && tc.calls >= 3 {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Retryable {
					t.Fatalf("unsafe retry: %v", err)
				}
			}
			want := [][]string{{"user", "info", "--uid", "tenant$user", "--format", "json"}, {"account", "get", "--account-id", id, "--format", "json"}, {"user", "modify", "--uid", "tenant$user", "--account-id=" + id, "--format", "json"}, {"user", "info", "--uid", "tenant$user", "--format", "json"}}
			for i, spec := range runner.specs {
				if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i == 2) || !reflect.DeepEqual(spec.Args, want[i]) {
					t.Fatalf("step %d: %+v", i, spec)
				}
			}
		})
	}
	for _, params := range []map[string]any{
		{"target_account_id": id}, {"target_account_id": id, "migration_confirm_uid": "other"}, {"migration_confirm_uid": "tenant$user"},
		{"target_account_id": " " + id, "migration_confirm_uid": "tenant$user"}, {"target_account_id": id, "migration_confirm_uid": "tenant$user", "account_root": true},
		{"target_account_id": id, "migration_confirm_uid": "tenant$user", "email": "new@example.test"},
	} {
		if _, err := build(Request{Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$user"}, params); err == nil {
			t.Fatalf("accepted invalid migration: %v", params)
		}
	}
}
