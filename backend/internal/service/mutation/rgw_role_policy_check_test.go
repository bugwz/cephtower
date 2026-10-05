package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRolePolicyAccountAndDuplicateChecks(t *testing.T) {
	p := map[string]any{"name": "reader", "account_id": "RGW123", "action": "put", "policy_name": "p", "policy_document": "{}"}
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{"RoleName":"reader","AccountId":"RGW123","PermissionPolicies":[{"PolicyName":"p","PolicyValue":"{}"}]}`, true},
		{`{"RoleName":"reader","AccountId":"","PermissionPolicies":[{"PolicyName":"p","PolicyValue":"{}"}]}`, false},
		{`{"RoleName":"reader","AccountId":"RGW123","PermissionPolicies":[{"PolicyName":"p","PolicyValue":"{}"},{"PolicyName":"p","PolicyValue":"{}"}]}`, false},
		{`{"RoleName":"reader","AccountId":"RGW123","PermissionPolicies":[null]}`, false},
	} {
		if rgwRolePolicyMatches(p, []byte(tc.raw)) != tc.valid {
			t.Fatalf("unexpected verification: %s", tc.raw)
		}
	}
}

func TestRolePolicyWriteVerifiesNativeState(t *testing.T) {
	for _, action := range []string{"put", "delete"} {
		for _, tc := range []struct {
			raw         string
			put, remove bool
		}{
			{`{"RoleName":"team$reader","AccountId":"","PermissionPolicies":[{"PolicyName":"p","PolicyValue":"{}"}]}`, true, false},
			{`{"RoleName":"team$reader","AccountId":""}`, false, true},
			{`{"RoleName":"team$reader","AccountId":"","PermissionPolicies":[]}`, false, true},
			{`{"RoleName":"team$reader","AccountId":"","PermissionPolicies":null}`, false, false},
			{`{"RoleName":"team$reader","AccountId":"","PermissionPolicies":[{"PolicyName":"p","PolicyValue":"wrong"}]}`, false, false},
			{`{"RoleName":"team$reader","AccountId":"","PermissionPolicies":[{"PolicyName":"other"}]}`, false, false},
			{`{"RoleName":"other","AccountId":""}`, false, false},
			{`{"RoleName":"team$reader","AccountId":"wrong"}`, false, false},
			{`{"RoleName":"team$reader"}`, false, false},
			{`null`, false, false}, {`invalid`, false, false},
		} {
			for _, fail := range []string{"", "rgw_role.policy", "rgw_role.policy.post_check"} {
				s, _, id := newCephUserService(t)
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_role.policy.post_check": tc.raw}, failID: fail}
				s.executor = runner
				_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_role.policy", Parameters: map[string]any{"name": "team$reader", "action": action, "policy_name": "p", "policy_document": "{}"}})
				valid := fail == "" && ((action == "put" && tc.put) || (action == "delete" && tc.remove))
				if (err == nil) != valid {
					t.Fatalf("%s %s %s: %v", action, tc.raw, fail, err)
				}
				if !valid {
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Retryable {
						t.Fatalf("unsafe failure: %v", err)
					}
				}
				count := 2
				if fail == "rgw_role.policy" {
					count = 1
				}
				if len(runner.specs) != count {
					t.Fatal("unexpected command continuation")
				}
			}
		}
	}
}
