package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"testing"
)

func TestSnapshotRetentionReadback(t *testing.T) {
	for _, tt := range []struct {
		name, action, post string
		valid              bool
	}{
		{"added", "add", `[{"retention":{"h":24,"d":7}}]`, true},
		{"unchanged", "add", `[{"retention":{"h":12,"d":7}}]`, false},
		{"missing", "add", `[{"retention":{"h":24}}]`, false},
		{"all schedules", "add", `[{"retention":{"h":24,"d":7}},{"retention":{"h":12,"d":7}}]`, false},
		{"removed", "remove", `[{"retention":{"n":3}}]`, true},
		{"still present", "remove", `[{"retention":{"h":24}}]`, false},
		{"empty", "remove", `[]`, false},
		{"missing policy", "remove", `[{}]`, false},
		{"trailing json", "add", `[{"retention":{"h":24,"d":7}}] {}`, false},
		{"wrong type", "add", `[{"retention":{"h":"24","d":7}}]`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"snapshot_schedule.retention.post_check": tt.post}}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "snapshot_schedule.retention", ResourceKey: "filesystem/data/snapshot-schedule", Parameters: map[string]any{"path": "/", "retention": "24h7d", "action": tt.action}})
			if tt.valid {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating {
				t.Fatalf("invalid command chain: %+v", runner.specs)
			}
		})
	}
}
