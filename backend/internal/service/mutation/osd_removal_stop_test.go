package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
)

type removalStopExecutor struct {
	specs         []executor.CommandSpec
	output, queue string
}

func (e *removalStopExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.Mutating {
		return executor.CommandResult{Stdout: []byte(e.output)}, nil
	}
	return executor.CommandResult{Stdout: []byte(e.queue)}, nil
}

func TestOSDRemovalStop(t *testing.T) {
	for _, tc := range []struct {
		output, queue string
		ok            bool
		calls         int
	}{
		{"Stopped OSD(s) removal", "No OSD remove/replace operations reported", true, 2},
		{"Stopped OSD(s) removal", `[{"osd_id":1}]`, true, 2},
		{"Stopped OSD(s) removal", `[{"osd_id":0}]`, false, 2},
		{"Stopped OSD(s) removal", `null`, false, 2},
		{"Stopped OSD(s) removal", `[{}]`, false, 2},
		{"Stopped OSD(s) removal", `[{"osd_id":1},{"osd_id":1}]`, false, 2},
		{"Unable to find OSD in the queue: 0", `[]`, false, 1},
	} {
		service, _, id := newCephUserService(t)
		runner := &removalStopExecutor{output: tc.output, queue: tc.queue}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd_removal.stop", ResourceKey: "osd-removal/0"})
		if (err == nil) != tc.ok || len(runner.specs) != tc.calls {
			t.Fatalf("%+v err=%v specs=%v", tc, err, runner.specs)
		}
		if !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "osd", "rm", "stop", "0"}) {
			t.Fatal(runner.specs)
		}
		if tc.calls == 2 && (runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, []string{"orch", "osd", "rm", "status", "--format", "json"})) {
			t.Fatal(runner.specs)
		}
	}
	for _, id := range []string{"all", "-1", "01", "2147483648"} {
		if _, err := build(Request{Action: "osd_removal.stop", ResourceKey: "osd-removal/" + id}, nil); err == nil {
			t.Fatal("invalid target accepted", id)
		}
	}
}
