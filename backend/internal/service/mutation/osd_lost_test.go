package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type lostExecutor struct{ destroyExecutor }

func (e *lostExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	spec.ID = strings.Replace(spec.ID, "osd.lost.", "osd.destroy.", 1)
	return e.destroyExecutor.Run(ctx, access, spec)
}

func TestOSDLostNativeEpochReadback(t *testing.T) {
	for _, scenario := range []string{"success", "unsafe", "up", "wrong-uuid", "missing-epoch", "unchanged", "new-down", "wrong-confirmation", "invalid-id", "command-failure", "read-failure"} {
		t.Run(scenario, func(t *testing.T) {
			service, _, cluster := newCephUserService(t)
			runner := &lostExecutor{destroyExecutor{before: `{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists"],"down_at":100,"lost_at":0}]}`, safety: `{"safe_to_destroy":[0],"active":[],"missing_stats":[],"stored_pgs":[]}`, after: `{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists"],"down_at":100,"lost_at":100}]}`}}
			service.executor = runner
			request := Request{ClusterID: cluster, Action: "osd.lost", ResourceKey: "osd/0", Parameters: map[string]any{"expected_uuid": "instance", "confirmation": "mark lost osd.0"}}
			calls := 4
			switch scenario {
			case "unsafe":
				runner.safety = `{"safe_to_destroy":[],"active":[0],"missing_stats":[],"stored_pgs":[]}`
				calls = 2
			case "up":
				runner.before = strings.Replace(runner.before, `"up":0`, `"up":1`, 1)
				calls = 1
			case "wrong-uuid":
				request.Parameters["expected_uuid"] = "other"
				calls = 1
			case "missing-epoch":
				runner.before = strings.Replace(runner.before, `"down_at":100,`, "", 1)
				calls = 1
			case "unchanged":
				runner.after = runner.before
			case "new-down":
				runner.after = strings.ReplaceAll(runner.after, "100", "101")
			case "wrong-confirmation":
				request.Parameters["confirmation"] = "mark lost osd.1"
				calls = 0
			case "invalid-id":
				request.ResourceKey = "osd/all"
				calls = 0
			case "command-failure":
				runner.failStage = "osd.destroy.execute"
				calls = 3
			case "read-failure":
				runner.failStage = "osd.destroy.post_check"
				runner.transportError = true
			}
			result, err := service.Execute(context.Background(), request)
			if (err == nil) != (scenario == "success") || len(runner.specs) != calls {
				t.Fatalf("err=%v commands=%v", err, runner.specs)
			}
			if scenario == "success" && result.Details.(map[string]any)["lost_at"] != uint32(100) {
				t.Fatal(result)
			}
			if scenario == "command-failure" || scenario == "read-failure" || scenario == "unchanged" || scenario == "new-down" {
				var actionError *cephdomain.ActionError
				if !errors.As(err, &actionError) || actionError.Retryable {
					t.Fatal(err)
				}
			}
			for i, spec := range runner.specs {
				if spec.Mutating != (i == 2) {
					t.Fatal(spec)
				}
				if i == 2 && !reflect.DeepEqual(spec.Args, []string{"osd", "lost", "0", "--yes-i-really-mean-it"}) {
					t.Fatal(spec)
				}
			}
		})
	}
}

func TestOSDLostEpochsStrict(t *testing.T) {
	for _, raw := range []string{`{}`, `{"osds":null}`, `{"osds":[]}`, `{"osds":[{"osd":0,"down_at":0,"lost_at":0}]}`, `{"osds":[{"osd":0,"down_at":4294967296,"lost_at":0}]}`, `{"osds":[{"osd":0,"down_at":1,"lost_at":null}]}`, `{"osds":[{"osd":0,"down_at":1,"lost_at":0},{"osd":0,"down_at":1,"lost_at":0}]}`} {
		if _, _, valid := osdLostEpochs([]byte(raw), "0"); valid {
			t.Fatal(raw)
		}
	}
	down, lost, valid := osdLostEpochs([]byte(`{"osds":[{"osd":0,"down_at":4294967295,"lost_at":4294967295}]}`), "0")
	if !valid || down != 4294967295 || lost != down {
		t.Fatal(down, lost, valid)
	}
}
