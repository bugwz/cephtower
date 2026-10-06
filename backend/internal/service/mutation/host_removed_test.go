package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type hostRemovalExecutor struct {
	specs []executor.CommandSpec
	raw   string
	exit  int
	err   error
}

func (e *hostRemovalExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.Mutating {
		return executor.CommandResult{}, nil
	}
	return executor.CommandResult{Stdout: []byte(e.raw), ExitCode: e.exit}, e.err
}

func TestHostRemovalReadback(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		exit  int
		err   error
		valid bool
	}{
		{`[]`, 0, nil, true}, {`[{"hostname":"node2"}]`, 0, nil, true},
		{`[{"hostname":"node1"}]`, 0, nil, false}, {`null`, 0, nil, false},
		{`{}`, 0, nil, false}, {`[{}]`, 0, nil, false}, {`[{"hostname":null}]`, 0, nil, false},
		{`[{"hostname":" "}]`, 0, nil, false}, {`[{"hostname":"node2"},{"hostname":"node2"}]`, 0, nil, false},
		{`[]`, 1, nil, false}, {`[]`, 0, errors.New("connection lost"), false},
	} {
		service, _, id := newCephUserService(t)
		runner := &hostRemovalExecutor{raw: tc.raw, exit: tc.exit, err: tc.err}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.delete", ResourceKey: "host/node1"})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
		if !tc.valid {
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Code != "post_check_failed" || actionError.Retryable {
				t.Fatal(err)
			}
		}
		if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating {
			t.Fatal(runner.specs)
		}
		if !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "host", "rm", "node1"}) || !reflect.DeepEqual(runner.specs[1].Args, []string{"orch", "host", "ls", "--detail", "--format", "json"}) {
			t.Fatal(runner.specs)
		}
	}
}
