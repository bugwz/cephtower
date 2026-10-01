package ceph

import (
	"context"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type upgradeStatusExecutor struct{ output string }

func (f upgradeStatusExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.ID == "collect.upgrade" {
		return executor.CommandResult{Stdout: []byte(f.output)}, nil
	}
	return fixtureExecutor{}.Run(ctx, access, spec)
}

func TestUpgradeCollectionRejectsUnknownState(t *testing.T) {
	for _, test := range []struct {
		output string
		valid  bool
	}{
		{`{"in_progress":false,"is_paused":false}`, true},
		{`{"in_progress":true,"is_paused":false}`, true},
		{`{"in_progress":true,"is_paused":true}`, true},
		{`null`, false},
		{`{}`, false},
		{`{"in_progress":false}`, false},
		{`{"in_progress":"false","is_paused":false}`, false},
		{`{"in_progress":false,"is_paused":true}`, false},
		{`{"in_progress":true,"is_paused":null}`, false},
	} {
		t.Run(test.output, func(t *testing.T) {
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			provider := NativeProvider{Executor: upgradeStatusExecutor{test.output}}
			rows, err := provider.collectTopology(ctx, ClusterAccess{})
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, row := range rows {
				if row.Kind == "upgrade" {
					found = true
				}
			}
			_, unavailable := trace.unavailable["upgrade"]
			if found != test.valid || unavailable == test.valid {
				t.Fatalf("found=%v unavailable=%v", found, unavailable)
			}
		})
	}
}
