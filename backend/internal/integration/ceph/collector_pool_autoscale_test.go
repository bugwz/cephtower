package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestPoolAutoscaleCollection(t *testing.T) {
	var calls []executor.CommandSpec
	p := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool_autoscale": []byte(`[{"pool_id":7,"pg_num_final":128,"would_adjust":false,"target_bytes":0,"subtree_capacity":4096,"target_ratio":0.5,"effective_target_ratio":0.25,"bias":1.5,"bulk":false},{"pool_id":8}]`),
	}}, calls: &calls}}
	got := p.collectPoolAutoscale(context.Background(), ClusterAccess{})
	if len(got) != 2 || got[7].PGNumFinal == nil || *got[7].PGNumFinal != 128 || got[7].WouldAdjust == nil || *got[7].WouldAdjust || got[7].TargetBytes == nil || *got[7].TargetBytes != 0 || got[8].WouldAdjust != nil {
		t.Fatalf("status = %#v", got)
	}
	if len(calls) != 1 || !reflect.DeepEqual(calls[0].Args, []string{"osd", "pool", "autoscale-status", "--format", "json"}) {
		t.Fatalf("commands = %#v", calls)
	}
	for _, response := range []string{`null`, `{}`, `[null]`, `[{}]`, `[{"pool_id":7},{"pool_id":7}]`, `[{"pool_id":7,"bias":-1}]`, `[{"pool_id":7,"pg_num_final":-1}]`, `[{"pool_id":7,"would_adjust":"false"}]`} {
		p.Executor = malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.pool_autoscale": []byte(response)}}
		if result := p.collectPoolAutoscale(context.Background(), ClusterAccess{}); result != nil {
			t.Fatalf("accepted %s", response)
		}
	}
	p.Executor = fixtureExecutor{t}
	if result := p.collectPoolAutoscale(context.Background(), ClusterAccess{}); result != nil {
		t.Fatal("unavailable command invented status")
	}
}

func TestPoolAutoscaleInventoryPayload(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool":           []byte(`[{"pool":7,"pool_name":"one","type":1},{"pool":8,"pool_name":"two","type":1}]`),
		"collect.pool_autoscale": []byte(`[{"pool_id":7,"pg_num_final":128,"would_adjust":true}]`),
	}}}
	rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, row := range rows {
		if row.Kind != "pool" {
			continue
		}
		found++
		data, err := json.Marshal(row.Payload)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatal(err)
		}
		if row.Name == "one" {
			status, ok := payload["autoscale_status"].(map[string]any)
			if !ok || status["pg_num_final"] != float64(128) || status["would_adjust"] != true {
				t.Fatalf("payload = %s", data)
			}
		} else if payload["autoscale_status"] != nil {
			t.Fatalf("cross-pool status: %s", data)
		}
	}
	if found != 2 {
		t.Fatalf("pool count %d", found)
	}
}

func TestPoolAutoscaleSizingInputs(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool_autoscale": []byte(`[{"pool_id":7,"logical_used":1024.5,"raw_used_rate":1.5,"actual_capacity_ratio":0.25,"capacity_ratio":1.25},{"pool_id":8,"logical_used":0,"actual_capacity_ratio":0}]`),
	}}}
	got := p.collectPoolAutoscale(context.Background(), ClusterAccess{})
	if len(got) != 2 {
		t.Fatalf("status = %#v", got)
	}
	data, err := json.Marshal(got[7])
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]float64{"logical_used": 1024.5, "raw_used_rate": 1.5, "actual_capacity_ratio": 0.25, "capacity_ratio": 1.25} {
		if decoded[key] != want {
			t.Fatalf("%s = %v, want %v", key, decoded[key], want)
		}
		p.Executor = malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.pool_autoscale": []byte(`[{"pool_id":7,"` + key + `":-1}]`)}}
		if bad := p.collectPoolAutoscale(context.Background(), ClusterAccess{}); bad != nil {
			t.Fatalf("accepted negative %s", key)
		}
	}
	if got[8].LogicalUsed == nil || *got[8].LogicalUsed != 0 || got[8].ActualCapacityRatio == nil || *got[8].ActualCapacityRatio != 0 || got[8].RawUsedRate != nil || got[8].CapacityRatio != nil {
		t.Fatalf("zero or unknown values lost: %#v", got[8])
	}
}
