package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type hostLabelExecutor struct {
	hostRemovalExecutor
	failFollowup bool
}

func (e *hostLabelExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if e.failFollowup && spec.ID == "host.update.step2" {
		e.specs = append(e.specs, spec)
		return executor.CommandResult{ExitCode: 1}, nil
	}
	return e.hostRemovalExecutor.Run(ctx, access, spec)
}

func TestHostLabelsReadback(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`[{"hostname":"node1","labels":["new","unrelated"]}]`, true},
		{`[{"hostname":"node1","labels":["new","old"]}]`, false},
		{`[{"hostname":"node1","labels":[]}]`, false},
		{`[{"hostname":"node1","labels":null}]`, false},
		{`[{"hostname":"node1","labels":["new","new"]}]`, false},
		{`[{"hostname":"node2","labels":["new"]}]`, false},
		{`[{"hostname":"node1","labels":["new"]},{"hostname":"node1","labels":["new"]}]`, false},
		{`null`, false}, {`{}`, false},
	} {
		service, _, id := newCephUserService(t)
		runner := &hostLabelExecutor{hostRemovalExecutor: hostRemovalExecutor{raw: tc.raw}}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.update", ResourceKey: "host/node1", Parameters: map[string]any{"labels_add": []any{"new"}, "labels_remove": []any{"old"}}})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
		if len(runner.specs) != 3 || runner.specs[2].Mutating {
			t.Fatal(runner.specs)
		}
	}
}

func TestHostLabelPartialFailureDoesNotRetry(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &hostLabelExecutor{failFollowup: true}
	service.executor = runner
	_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.update", ResourceKey: "host/node1", Parameters: map[string]any{"labels_add": []any{"new"}, "labels_remove": []any{"old"}}})
	var actionError *cephdomain.ActionError
	if !errors.As(err, &actionError) || actionError.Retryable || len(runner.specs) != 2 {
		t.Fatal(err, runner.specs)
	}
}

func TestHostLabelInvalidChanges(t *testing.T) {
	for _, p := range []map[string]any{
		{"labels_add": []any{""}}, {"labels_add": []any{" new"}}, {"labels_add": []any{"--force"}},
		{"labels_add": []any{"new", "new"}}, {"labels_add": []any{"new"}, "labels_remove": []any{"new"}},
	} {
		if _, err := build(Request{Action: "host.update", ResourceKey: "host/node1"}, p); err == nil {
			t.Fatal(p)
		}
	}
}
