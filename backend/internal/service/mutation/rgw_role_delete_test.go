package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWRoleDeletionVerifiesScopedAbsence(t *testing.T) {
	for _, account := range []string{"", "RGW123"} {
		for _, tc := range []struct {
			raw, fail string
			valid     bool
		}{
			{`[]`, "", true},
			{`[{"RoleName":"other","AccountId":"` + account + `"}]`, "", true},
			{`[{"RoleName":"team$reader","AccountId":"` + account + `"}]`, "", false},
			{`[{"RoleName":"other","AccountId":"wrong"}]`, "", false},
			{`[{"RoleName":"other"}]`, "", false},
			{`null`, "", false}, {`{}`, "", false}, {`[null]`, "", false},
			{`[{"RoleName":"other","AccountId":"` + account + `"},{"RoleName":"other","AccountId":"` + account + `"}]`, "", false},
			{`[] []`, "", false}, {`invalid`, "", false},
			{`{"Roles":[],"next-marker":"next"}`, "", false},
			{`[]`, "rgw_role.delete", false}, {`[]`, "rgw_role.delete.post_check", false},
		} {
			s, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_role.delete.post_check": tc.raw}, failID: tc.fail}
			s.executor = runner
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_role.delete", Parameters: map[string]any{"name": "team$reader", "account_id": account}})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
			if !tc.valid {
				var failure *cephdomain.ActionError
				code := "post_check_failed"
				if tc.fail == "rgw_role.delete" {
					code = "ceph_command_failed"
				}
				if !errors.As(err, &failure) || failure.Code != code || failure.Retryable {
					t.Fatalf("unsafe failure: %v", err)
				}
			}
			count := 2
			if tc.fail == "rgw_role.delete" {
				count = 1
			}
			if len(runner.specs) != count {
				t.Fatal("unexpected continuation")
			}
			if count == 2 {
				args := []string{"role", "list", "--tenant", "team", "--format", "json"}
				if account != "" {
					args = []string{"role", "list", "--account-id", account, "--format", "json"}
				}
				if runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, args) {
					t.Fatalf("wrong read: %+v", runner.specs[1])
				}
			}
		}
	}
}

func TestRGWRoleNameValidation(t *testing.T) {
	for _, name := range []any{"", "$reader", "team$", "a$b$c", " reader", "reader\n", "a;rm", 123, nil} {
		if _, err := rgwRoleName(map[string]any{"name": name}); err == nil {
			t.Fatalf("accepted invalid name: %v", name)
		}
	}
	for _, name := range []string{"reader", "team$reader"} {
		if got, err := rgwRoleName(map[string]any{"name": name}); err != nil || got != name {
			t.Fatalf("rejected name %q: %v", name, err)
		}
	}
}
