package ceph

import (
	"context"
	"strings"
	"testing"
)

func TestMonitorInventoryRequiresExplicitUniqueNames(t *testing.T) {
	for _, raw := range []string{`{}`, `null`, `{"mons":null}`, `{"mons":[null]}`, `{"mons":[{}]}`, `{"mons":[{"name":" a"}]}`, `{"mons":[{"name":"a "}]}`, `{"mons":[{"name":"a"},{"name":"a"}]}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mon": []byte(raw)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err == nil || !strings.Contains(err.Error(), "collect.mon") || len(rows) != 0 {
			t.Fatalf("accepted invalid inventory %s: rows=%d err=%v", raw, len(rows), err)
		}
	}
	for _, raw := range []string{`{"mons":[]}`, `{"mons":[{"name":"a","rank":0},{"name":"b","rank":1}]}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mon": []byte(raw)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, row := range rows {
			if row.Kind == "mon" {
				count++
			}
		}
		want := 2
		if raw == `{"mons":[]}` {
			want = 0
		}
		if count != want {
			t.Fatalf("count=%d want=%d", count, want)
		}
	}
}
