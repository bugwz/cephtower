package ceph

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

type erasureProfileExecutor struct{ list, detail string }

func (e erasureProfileExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	data := e.list
	if spec.ID == "collect.erasure_code_profile_detail" {
		data = e.detail
	}
	if data == "" {
		return executor.CommandResult{}, fmt.Errorf("read failed")
	}
	return executor.CommandResult{Stdout: []byte(data)}, nil
}

func TestErasureProfileCollectionCompleteness(t *testing.T) {
	for _, tc := range []struct {
		list, detail string
		valid        bool
		count        int
	}{
		{`[]`, "", true, 0},
		{`["ec"]`, `{"plugin":"isa","k":"4","m":"2"}`, true, 1},
		{`null`, "", false, 0},
		{`[""]`, "", false, 0},
		{`["ec","ec"]`, `{"plugin":"isa"}`, false, 0},
		{`["ec"]`, "", false, 0},
		{`["ec"]`, `null`, false, 0},
		{`["ec"]`, `{}`, false, 0},
		{`["ec"]`, `{"plugin":2}`, false, 0},
		{`["ec"]`, `{"plugin":"isa","k":4}`, false, 0},
	} {
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		p := NativeProvider{Executor: erasureProfileExecutor{tc.list, tc.detail}}
		rows := p.collectErasureProfiles(ctx, ClusterAccess{}, time.Now())
		_, unavailable := trace.unavailable["erasure_code_profile"]
		if len(rows) != tc.count || unavailable == tc.valid {
			t.Fatalf("%+v: rows=%v unavailable=%v", tc, rows, unavailable)
		}
	}
}
