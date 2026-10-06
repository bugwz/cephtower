package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type hostRescanExecutor struct {
	hostRemovalExecutor
	ack string
}

func (e *hostRescanExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.Mutating {
		e.specs = append(e.specs, spec)
		return executor.CommandResult{Stdout: []byte(e.ack)}, nil
	}
	return e.hostRemovalExecutor.Run(ctx, access, spec)
}

func TestHostRescanNativeAcknowledgement(t *testing.T) {
	for _, tc := range []struct {
		ack   string
		valid bool
	}{
		{"Ok. No compatible HBAs found", true},
		{"Ok. 3 adapters detected: 2 rescanned, 1 skipped, 0 failed (0.12s)\n", true},
		{"Partial. 1 successful, 1 failure against: host0", false},
		{"Failed. All 2 rescan requests failed", false},
		{"", false}, {"OK", false}, {"Ok. done\nFailed. scan", false},
	} {
		service, _, id := newCephUserService(t)
		runner := &hostRescanExecutor{ack: tc.ack, hostRemovalExecutor: hostRemovalExecutor{raw: `[{"hostname":"node1"}]`}}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.action", ResourceKey: "host/node1/action", Parameters: map[string]any{"action": "rescan"}})
		if (err == nil) != tc.valid {
			t.Fatalf("ack=%q err=%v", tc.ack, err)
		}
		if !tc.valid {
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Code != "post_check_failed" || actionError.Retryable || len(runner.specs) != 1 {
				t.Fatalf("ack=%q err=%v commands=%v", tc.ack, err, runner.specs)
			}
		}
	}
}
