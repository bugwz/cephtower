package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWAccountDeletionRequiresVerifiedAbsence(t *testing.T) {
	for _, tc := range []struct {
		name, raw, failID string
		success           bool
	}{
		{"empty", `[]`, "", true},
		{"other account", `["RGW1234","RGWother"]`, "", true},
		{"retained", `["RGW123"]`, "", false},
		{"null", `null`, "", false},
		{"object", `{}`, "", false},
		{"page", `{"keys":[],"truncated":true}`, "", false},
		{"page false", `{"keys":[],"truncated":false}`, "", false},
		{"null member", `[null]`, "", false},
		{"empty key", `[""]`, "", false},
		{"numeric key", `[1]`, "", false},
		{"truncated json", `["other"`, "", false},
		{"trailing document", `[] []`, "", false},
		{"read failure", `[]`, "rgw_account.delete.post_check", false},
		{"write failure", `[]`, "rgw_account.delete", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_account.delete.post_check": tc.raw}, failID: tc.failID}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_account.delete", Parameters: map[string]any{"account_id": "RGW123"}})
			if (err == nil) != tc.success {
				t.Fatalf("success=%t err=%v", tc.success, err)
			}
			if !tc.success {
				var failure *cephdomain.ActionError
				code := "post_check_failed"
				if tc.failID == "rgw_account.delete" {
					code = "ceph_command_failed"
				}
				if !errors.As(err, &failure) || failure.Code != code || failure.Retryable {
					t.Fatalf("unsafe failure: %v", err)
				}
			}
			count := 2
			if tc.failID == "rgw_account.delete" {
				count = 1
			}
			if len(runner.specs) != count {
				t.Fatalf("commands=%v", runner.specs)
			}
			for i, spec := range runner.specs {
				want := []string{"account", "rm", "--account-id", "RGW123", "--format", "json"}
				if i == 1 {
					want = []string{"account", "list", "--format", "json"}
				}
				if !reflect.DeepEqual(spec.Args, want) || spec.Mutating != (i == 0) {
					t.Fatalf("unexpected command: %v", spec)
				}
			}
		})
	}
}
