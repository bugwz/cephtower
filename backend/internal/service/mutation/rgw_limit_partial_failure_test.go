package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type limitPartialFailureExecutor struct {
	specs   []executor.CommandSpec
	failure error
}

func (e *limitPartialFailureExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if len(e.specs) == 2 {
		return executor.CommandResult{}, e.failure
	}
	return executor.CommandResult{}, nil
}

func TestRGWLimitPartialFailureStopsWithoutRetry(t *testing.T) {
	for _, tc := range []struct {
		action, scope string
		enabled       bool
	}{
		{"rgw_user.quota", "user", false},
		{"rgw_user.quota", "bucket", false},
		{"rgw_bucket.quota", "bucket", false},
		{"rgw_user.ratelimit", "user", true},
		{"rgw_user.ratelimit", "user", false},
		{"rgw_bucket.ratelimit", "bucket", true},
		{"rgw_bucket.ratelimit", "bucket", false},
	} {
		for _, failure := range []error{errors.New("private transport diagnostic"), context.DeadlineExceeded, context.Canceled} {
			service, _, id := newCephUserService(t)
			runner := &limitPartialFailureExecutor{failure: failure}
			service.executor = runner
			params := map[string]any{"uid": "team$user", "bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00photos")), "scope": tc.scope, "enabled": tc.enabled,
				"max_size": json.Number("1025"), "max_objects": json.Number("0"), "max_read_ops": json.Number("0"), "max_write_ops": json.Number("1"), "max_read_bytes": json.Number("1024"), "max_write_bytes": json.Number("2048")}
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: tc.action, Parameters: params})
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != "ceph_command_failed" || actionErr.Retryable {
				t.Fatalf("%s: expected non-retryable partial failure, got %v", tc.action, err)
			}
			if !strings.Contains(actionErr.Message, "may already have changed") || !strings.Contains(actionErr.Message, "inspect current limits and activation") || strings.Contains(actionErr.Message, failure.Error()) {
				t.Fatalf("missing partial-state guidance or leaked diagnostic: %s", actionErr.Message)
			}
			verb := "disable"
			if tc.enabled {
				verb = "enable"
			}
			if len(runner.specs) != 2 || !runner.specs[0].Mutating || !runner.specs[1].Mutating || runner.specs[0].Args[1] != "set" || runner.specs[1].Args[1] != verb {
				t.Fatalf("unexpected command continuation: %#v", runner.specs)
			}
		}
	}
}
