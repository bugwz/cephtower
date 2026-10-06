package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type destroyExecutor struct {
	specs                 []executor.CommandSpec
	before, safety, after string
	failStage             string
	transportError        bool
}

func (e *destroyExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.ID == e.failStage {
		if e.transportError {
			return executor.CommandResult{}, errors.New("connection lost")
		}
		return executor.CommandResult{ExitCode: 1}, nil
	}
	output := e.after
	if spec.ID == "osd.destroy.pre_check" {
		output = e.before
	}
	if spec.ID == "osd.destroy.safety" {
		output = e.safety
	}
	return executor.CommandResult{Stdout: []byte(output)}, nil
}

func TestOSDDestroyCommandFailuresNeverRetry(t *testing.T) {
	for index, stage := range []string{"pre_check", "safety", "execute", "post_check"} {
		for _, transportError := range []bool{false, true} {
			service, _, cluster := newCephUserService(t)
			runner := &destroyExecutor{
				before:    `{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists"]}]}`,
				safety:    `{"safe_to_destroy":[0],"active":[],"missing_stats":[],"stored_pgs":[]}`,
				failStage: "osd.destroy." + stage, transportError: transportError,
			}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "osd.destroy", ResourceKey: "osd/0", Parameters: map[string]any{"expected_uuid": "instance", "confirmation": "destroy osd.0"}})
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Retryable || len(runner.specs) != index+1 {
				t.Fatalf("%s transport=%v err=%v commands=%v", stage, transportError, err, runner.specs)
			}
		}
	}
}

func TestOSDDestroyRejectsAmbiguousState(t *testing.T) {
	for _, raw := range []string{
		`null`, `{}`, `{"osds":[]}`,
		`{"osds":[{"osd":0,"uuid":"instance","up":0,"state":[]}]}`,
		`{"osds":[{"osd":0,"uuid":"instance","up":0,"state":[null]}]}`,
		`{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists","exists"]}]}`,
		`{"osds":[{"osd":0,"uuid":"instance","up":0,"state":[" destroyed "]}]}`,
		`{"osds":[{"osd":0,"uuid":"instance","state":["exists"]}]}`,
		`{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["destroyed"]}]}`,
		`{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists"]},{"osd":0,"uuid":"instance","up":0,"state":["exists"]}]}`,
	} {
		if osdDestroyState([]byte(raw), "0", "instance", false) {
			t.Fatalf("accepted ambiguous state: %s", raw)
		}
	}
}

func TestOSDDestroySafetyAndReadback(t *testing.T) {
	for _, scenario := range []string{"success", "up", "identity", "unsafe", "invalid-safety", "not-destroyed", "old-uuid", "confirmation", "invalid-id"} {
		service, _, cluster := newCephUserService(t)
		runner := &destroyExecutor{before: `{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists"]}]}`, safety: `{"safe_to_destroy":[0],"active":[],"missing_stats":[],"stored_pgs":[]}`, after: `{"osds":[{"osd":0,"uuid":"00000000-0000-0000-0000-000000000000","up":0,"state":["destroyed"]}]}`}
		service.executor = runner
		request := Request{ClusterID: cluster, Action: "osd.destroy", ResourceKey: "osd/0", Parameters: map[string]any{"expected_uuid": "instance", "confirmation": "destroy osd.0"}}
		calls := 4
		switch scenario {
		case "up":
			runner.before = `{"osds":[{"osd":0,"uuid":"instance","up":1,"state":["exists"]}]}`
			calls = 1
		case "identity":
			request.Parameters["expected_uuid"] = "other"
			calls = 1
		case "unsafe":
			runner.safety = `{"safe_to_destroy":[],"active":[0],"missing_stats":[],"stored_pgs":[]}`
			calls = 2
		case "invalid-safety":
			runner.safety = `{}`
			calls = 2
		case "not-destroyed":
			runner.after = runner.before
		case "old-uuid":
			runner.after = `{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["destroyed"]}]}`
		case "confirmation":
			request.Parameters["confirmation"] = "destroy osd.1"
			calls = 0
		case "invalid-id":
			request.ResourceKey = "osd/all"
			calls = 0
		}
		_, err := service.Execute(context.Background(), request)
		if (err == nil) != (scenario == "success") || len(runner.specs) != calls {
			t.Fatalf("%s err=%v specs=%v", scenario, err, runner.specs)
		}
		for i, spec := range runner.specs {
			if spec.Mutating != (i == 2) {
				t.Fatal(spec)
			}
			if i == 2 && !reflect.DeepEqual(spec.Args, []string{"osd", "destroy-actual", "0", "--yes-i-really-mean-it"}) {
				t.Fatal(spec)
			}
		}
	}
}
