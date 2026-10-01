package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestPoolIOCollection(t *testing.T) {
	var calls []executor.CommandSpec
	p := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool_io": []byte(`[{"pool_id":7,"client_io_rate":{"read_bytes_sec":1024,"write_bytes_sec":0,"read_op_per_sec":2,"write_op_per_sec":0}},{"pool_id":8,"client_io_rate":{}},{"pool_id":9}]`),
	}}, calls: &calls}}
	got := p.collectPoolIO(context.Background(), ClusterAccess{})
	if len(got) != 3 || got[7] == nil || got[8] == nil || got[9] != nil {
		t.Fatalf("rates = %#v", got)
	}
	data, err := json.Marshal(got[7])
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]float64{"read_bytes_sec": 1024, "write_bytes_sec": 0, "read_op_per_sec": 2, "write_op_per_sec": 0} {
		if decoded[key] != want {
			t.Fatalf("rate %s = %v", key, decoded[key])
		}
	}
	if got[8].ReadBytesSec != nil || got[8].WriteOpsSec != nil {
		t.Fatal("empty rates invented zero")
	}
	if len(calls) != 1 || !reflect.DeepEqual(calls[0].Args, []string{"osd", "pool", "stats", "--format", "json"}) {
		t.Fatalf("commands = %#v", calls)
	}
	for _, response := range []string{`null`, `{}`, `[null]`, `[{}]`, `[{"pool_id":1},{"pool_id":1}]`, `[{"pool_id":1,"client_io_rate":{"read_bytes_sec":-1}}]`, `[{"pool_id":1,"client_io_rate":{"write_op_per_sec":1.5}}]`} {
		p.Executor = malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.pool_io": []byte(response)}}
		if result := p.collectPoolIO(context.Background(), ClusterAccess{}); result != nil {
			t.Fatalf("accepted %s", response)
		}
	}
	p.Executor = fixtureExecutor{t}
	if p.collectPoolIO(context.Background(), ClusterAccess{}) != nil {
		t.Fatal("unavailable rates")
	}
}

func TestPoolIORatesReachInventory(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool":    []byte(`[{"pool":7,"pool_name":"io","type":1}]`),
		"collect.pool_io": []byte(`[{"pool_id":7,"client_io_rate":{"read_bytes_sec":4096}},{"pool_id":8,"client_io_rate":{"read_bytes_sec":8192}}]`),
	}}}
	rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Kind != "pool" {
			continue
		}
		data, err := json.Marshal(row.Payload)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatal(err)
		}
		rate, ok := payload["client_io_rate"].(map[string]any)
		if !ok || rate["read_bytes_sec"] != float64(4096) || rate["write_bytes_sec"] != nil {
			t.Fatalf("payload = %s", data)
		}
		return
	}
	t.Fatal("missing pool")
}
