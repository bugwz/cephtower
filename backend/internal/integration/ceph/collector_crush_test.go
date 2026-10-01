package ceph

import (
	"context"
	"testing"
	"time"
)

func TestCrushRuleCollectionIdentity(t *testing.T) {
	for _, tc := range []struct {
		data  string
		valid bool
		count int
	}{
		{`[]`, true, 0},
		{`[{"rule_id":8,"rule_name":"ssd","steps":[]},{"rule_id":0,"rule_name":"hdd"}]`, true, 2},
		{`null`, false, 0}, {`{}`, false, 0}, {`[null]`, false, 0},
		{`[{"rule_id":0}]`, false, 0},
		{`[{"rule_id":"0","rule_name":"x"}]`, false, 0},
		{`[{"rule_id":-1,"rule_name":"x"}]`, false, 0},
		{`[{"rule_id":0,"rule_name":"x"},{"rule_id":0,"rule_name":"y"}]`, false, 0},
		{`[{"rule_id":0,"rule_name":"x"},{"rule_id":1,"rule_name":"x"}]`, false, 0},
	} {
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		p := NativeProvider{Executor: erasureProfileExecutor{list: tc.data}}
		rows := p.collectCrushRules(ctx, ClusterAccess{}, time.Now())
		_, unavailable := trace.unavailable["crush_rule"]
		if len(rows) != tc.count || unavailable == tc.valid {
			t.Fatalf("%s: rows=%v unavailable=%v", tc.data, rows, unavailable)
		}
	}
}
