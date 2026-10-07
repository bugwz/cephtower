package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestMirrorLeaderCounts(t *testing.T) {
	const raw = `{"7":{"name":"pool","leader":true,"instance_id":"instance","namespaces":{"":{"image_local_count":0,"image_remote_count":18446744073709551615},"ns":{"image_local_count":-1,"image_remote_count":"9"}}}}`
	leader := map[string]any{"leader": true, "service_id": "svc", "instance_id": "instance"}
	for _, scenario := range []string{"success", "no_leader", "multiple", "wrong_pool", "wrong_instance", "malformed", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			pool := poolWire{Pool: 7, PoolName: "pool"}
			var list any = []any{leader}
			status := map[string]any{"rbd-mirror": map[string]any{"svc": map[string]any{"status": map[string]any{"json": raw}}}}
			switch scenario {
			case "no_leader":
				list = []any{}
			case "multiple":
				list = []any{leader, leader}
			case "wrong_pool":
				pool.Pool = 8
			case "wrong_instance":
				list = []any{map[string]any{"leader": true, "service_id": "svc", "instance_id": "other"}}
			case "malformed":
				status["rbd-mirror"] = false
			case "missing":
				status = nil
			}
			got := mirrorLeaderCounts(status, pool, list)
			if scenario != "success" {
				if got != nil {
					t.Fatalf("unexpected counts: %#v", got)
				}
				return
			}
			want := []map[string]any{{"namespace": "", "image_local_count": "0", "image_remote_count": "18446744073709551615"}, {"namespace": "ns"}}
			if !reflect.DeepEqual(got["namespaces"], want) {
				t.Fatalf("counts=%#v", got)
			}
		})
	}
}

func TestMirrorLeaderCountsCollection(t *testing.T) {
	report := `{"7":{"name":"pool","leader":true,"instance_id":"instance","namespaces":{"":{"image_local_count":0}}}}`
	raw, _ := json.Marshal(map[string]any{"rbd-mirror": map[string]any{"svc": map[string]any{"status": map[string]any{"json": report}}}})
	var calls []executor.CommandSpec
	p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_mirroring":             []byte(`{"mode":"image","peers":[]}`),
		"collect.rbd_mirroring_status":      []byte(`{"daemons":[{"leader":true,"service_id":"svc","instance_id":"instance"}]}`),
		"collect.rbd_mirror_service_status": raw,
	}}}}
	rows := p.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{Pool: 7, PoolName: "pool"}, {Pool: 8, PoolName: "other"}}, fsDumpWire{}, time.Now())
	found := false
	for _, row := range rows {
		if row.Kind != "rbd_mirroring" {
			continue
		}
		counts := row.Payload.(map[string]any)["leader_counts"]
		if row.NaturalKey == "pool" {
			found = counts != nil
		} else if counts != nil {
			t.Fatal("cross-pool counts")
		}
	}
	if !found {
		t.Fatal("missing counts")
	}
	n := 0
	for _, call := range calls {
		if call.ID == "collect.rbd_mirror_service_status" {
			n++
			if call.Binary != executor.BinaryCeph || call.Mutating || !reflect.DeepEqual(call.Args, []string{"service", "status", "--format", "json"}) {
				t.Fatalf("command=%#v", call)
			}
		}
	}
	if n != 1 {
		t.Fatalf("service status calls=%d", n)
	}
}
