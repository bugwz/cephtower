package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestHostActionStateReadback(t *testing.T) {
	for _, tc := range []struct {
		action, raw string
		valid       bool
	}{
		{"maintenance_enter", `[{"hostname":"node1","status":"maintenance"}]`, true},
		{"maintenance_exit", `[{"hostname":"node1","status":""}]`, true},
		{"maintenance_exit", `[{"hostname":"node1","status":"offline"}]`, false},
		{"maintenance_enter", `[{"hostname":"node1","status":null}]`, false},
		{"maintenance_enter", `[{"hostname":"node2","status":"maintenance"}]`, false},
		{"maintenance_enter", `[{"hostname":"node1","status":"maintenance"},{"hostname":"node1","status":"maintenance"}]`, false},
		{"drain", `[{"hostname":"node1","labels":["_no_schedule","_no_conf_keyring","custom"]}]`, true},
		{"drain", `[{"hostname":"node1","labels":["_no_schedule"]}]`, false},
		{"drain", `[{"hostname":"node1","labels":["_no_schedule","_no_conf_keyring",null]}]`, false},
		{"stop_drain", `[{"hostname":"node1","labels":[]}]`, true},
		{"stop_drain", `[{"hostname":"node1","labels":["custom"]}]`, true},
		{"stop_drain", `[{"hostname":"node1","labels":["_no_conf_keyring"]}]`, false},
		{"stop_drain", `[{"hostname":"node1"}]`, false},
		{"stop_drain", `null`, false}, {"maintenance_enter", `[]`, false},
	} {
		service, _, id := newCephUserService(t)
		runner := &hostRemovalExecutor{raw: tc.raw}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.action", ResourceKey: "host/node1/action", Parameters: map[string]any{"action": tc.action}})
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
	}
}
