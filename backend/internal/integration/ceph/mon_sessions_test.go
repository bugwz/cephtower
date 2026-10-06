package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestMonitorSessionsPreserveUnsignedCounts(t *testing.T) {
	for _, raw := range []string{"0", "15", "9007199254740993", "18446744073709551615", "18446744073709551616", "-1", "1.5", `"15"`, "null", "false", "{}"} {
		t.Run(raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.mon_perf.schema": []byte(`{"mon":{"num_sessions":{"type":2,"priority":5,"value_type":"integer"}}}`),
				"collect.mon_perf.dump":   []byte(`{"mon":{"num_sessions":` + raw + `}}`),
			}}}
			rows, count := p.collectMonitorPerfCounters(context.Background(), ClusterAccess{}, "a", 5, time.Now())
			valid := raw == "0" || raw == "15" || raw == "9007199254740993" || raw == "18446744073709551615"
			if (count != nil) != valid || (valid && *count != raw) {
				t.Fatalf("count=%v", count)
			}
			if len(rows) != 1 {
				t.Fatal("missing counter")
			}
			encoded, _ := json.Marshal(rows[0].Payload)
			var payload map[string]any
			if json.Unmarshal(encoded, &payload) != nil {
				t.Fatal("invalid payload")
			}
			for _, field := range []string{"value", "raw_value"} {
				if valid && payload[field] != raw {
					t.Fatalf("lost precision: %s", encoded)
				}
				if !valid && payload[field] != nil {
					t.Fatalf("invented count: %s", encoded)
				}
			}
		})
	}
}

func TestMonitorSessionsIndependentOfPerfThreshold(t *testing.T) {
	for _, schema := range []string{
		`{"mon":{"num_sessions":{"type":2,"priority":5,"value_type":"integer"}}}`,
		`{}`,
	} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.mon_perf.schema": []byte(schema),
			"collect.mon_perf.dump":   []byte(`{"mon":{"num_sessions":9007199254740993}}`),
		}}}
		rows, count := p.collectMonitorPerfCounters(context.Background(), ClusterAccess{}, "a", 10, time.Now())
		if len(rows) != 0 {
			t.Fatal("performance threshold was ignored")
		}
		if count == nil || *count != "9007199254740993" {
			t.Fatalf("lost independent session count: %v", count)
		}
	}
}
