package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
)

type rateLimitExecutor struct {
	specs    []executor.CommandSpec
	response []byte
}

func (e *rateLimitExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.Mutating {
		return executor.CommandResult{}, nil
	}
	return executor.CommandResult{Stdout: e.response}, nil
}

func TestRateLimitExecutesFinalReadback(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &rateLimitExecutor{response: []byte(`{"user_ratelimit":{"enabled":true,"max_read_ops":0,"max_write_ops":0,"max_read_bytes":0,"max_write_bytes":0}}`)}
	service.executor = runner
	_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_user.ratelimit", Parameters: map[string]any{"uid": "user", "enabled": true, "max_read_ops": float64(0), "max_write_ops": float64(0), "max_read_bytes": float64(0), "max_write_bytes": float64(0)}})
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 3 {
		t.Fatalf("expected set, enable, get; got %#v", runner.specs)
	}
	check := runner.specs[2]
	if check.Mutating || !reflect.DeepEqual(check.Args, []string{"ratelimit", "get", "--uid", "user", "--ratelimit-scope", "user", "--format", "json"}) {
		t.Fatalf("wrong final check: %#v", check)
	}
}

func TestEmptyRGWCheckIsNotACommand(t *testing.T) {
	cmd, err := build(Request{Action: "rgw_user.ratelimit"}, map[string]any{"uid": "user", "enabled": false, "max_read_ops": float64(0), "max_write_ops": float64(0), "max_read_bytes": float64(0), "max_write_bytes": float64(0)})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.check) != 0 {
		t.Fatalf("unexpected initial check: %#v", cmd.check)
	}
}
