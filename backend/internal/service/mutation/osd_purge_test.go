package mutation

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type purgeExecutor struct {
	destroyExecutor
	crush string
}

func (e *purgeExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.ID == "osd.purge.crush_check" {
		e.specs = append(e.specs, spec)
		return executor.CommandResult{Stdout: []byte(e.crush)}, nil
	}
	spec.ID = strings.Replace(spec.ID, "osd.purge.", "osd.destroy.", 1)
	return e.destroyExecutor.Run(ctx, access, spec)
}

func TestOSDPurgeSafetyAndReadback(t *testing.T) {
	for _, scenario := range []string{"success", "destroyed", "identity", "up", "unsafe", "still-present", "invalid-map", "crush-present", "invalid-crush", "confirmation", "invalid-id", "command-failure"} {
		t.Run(scenario, func(t *testing.T) {
			service, _, cluster := newCephUserService(t)
			runner := &purgeExecutor{destroyExecutor: destroyExecutor{before: `{"osds":[{"osd":0,"uuid":"instance","up":0,"state":["exists"]}]}`, safety: `{"safe_to_destroy":[0],"active":[],"missing_stats":[],"stored_pgs":[]}`, after: `{"osds":[]}`}, crush: `{"devices":[{"id":0,"name":"device0"}],"buckets":[]}`}
			service.executor = runner
			request := Request{ClusterID: cluster, Action: "osd.purge", ResourceKey: "osd/0", Parameters: map[string]any{"expected_uuid": "instance", "confirmation": "purge osd.0"}}
			calls := 5
			switch scenario {
			case "destroyed":
				request.Parameters["expected_uuid"] = "00000000-0000-0000-0000-000000000000"
				runner.before = `{"osds":[{"osd":0,"uuid":"00000000-0000-0000-0000-000000000000","up":0,"state":["destroyed"]}]}`
			case "identity":
				request.Parameters["expected_uuid"] = "other"
				calls = 1
			case "up":
				runner.before = strings.Replace(runner.before, `"up":0`, `"up":1`, 1)
				calls = 1
			case "unsafe":
				runner.safety = `{"safe_to_destroy":[],"active":[0],"missing_stats":[],"stored_pgs":[]}`
				calls = 2
			case "still-present":
				runner.after = runner.before
				calls = 4
			case "invalid-map":
				runner.after = `{}`
				calls = 4
			case "crush-present":
				runner.crush = `{"devices":[{"id":0,"name":"osd.0"}],"buckets":[]}`
			case "invalid-crush":
				runner.crush = `{}`
			case "confirmation":
				request.Parameters["confirmation"] = "purge osd.1"
				calls = 0
			case "invalid-id":
				request.ResourceKey = "osd/all"
				calls = 0
			case "command-failure":
				runner.failStage = "osd.destroy.execute"
				calls = 3
			}
			_, err := service.Execute(context.Background(), request)
			if (err == nil) != (scenario == "success" || scenario == "destroyed") || len(runner.specs) != calls {
				t.Fatalf("err=%v specs=%v", err, runner.specs)
			}
			for i, spec := range runner.specs {
				if spec.Mutating != (i == 2) {
					t.Fatal(spec)
				}
				if i == 2 && !reflect.DeepEqual(spec.Args, []string{"osd", "purge-actual", "0", "--yes-i-really-mean-it"}) {
					t.Fatal(spec)
				}
			}
		})
	}
}

func TestOSDPurgeCrushPlaceholders(t *testing.T) {
	for _, raw := range []string{`{"devices":[],"buckets":[]}`, `{"devices":[{"id":0,"name":"device0"}],"buckets":[{"items":[{"id":1}]}]}`} {
		if !osdPurgeCrushRemoved([]byte(raw), "0") {
			t.Fatal(raw)
		}
	}
	for _, raw := range []string{`{}`, `{"devices":[],"buckets":null}`, `{"devices":[],"buckets":[{}]}`, `{"devices":[{"id":0,"name":"osd.0"}],"buckets":[]}`, `{"devices":[{"id":0,"name":"device0"}],"buckets":[{"items":[{"id":0}]}]}`, `{"devices":[{"id":0,"name":"device0"},{"id":0,"name":"device0"}],"buckets":[]}`} {
		if osdPurgeCrushRemoved([]byte(raw), "0") {
			t.Fatal(raw)
		}
	}
}

func TestOSDPurgeAbsentStrict(t *testing.T) {
	for _, raw := range []string{`{}`, `null`, `{"devices":null}`, `{"devices":[{}]}`, `{"devices":[{"id":null}]}`, `{"devices":[{"id":-1}]}`, `{"devices":[{"id":2147483648}]}`, `{"devices":[{"id":1},{"id":1}]}`, `{"devices":[{"id":0}]}`} {
		if osdPurgeAbsent([]byte(raw), "devices", "id", "0") {
			t.Fatal(raw)
		}
	}
	for _, raw := range []string{`{"devices":[]}`, `{"devices":[{"id":1}]}`} {
		if !osdPurgeAbsent([]byte(raw), "devices", "id", "0") {
			t.Fatal(raw)
		}
	}
}
