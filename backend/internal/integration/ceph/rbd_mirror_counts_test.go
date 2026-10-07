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
		status := row.Payload.(map[string]any)["leader_counts_status"]
		if row.NaturalKey == "pool" {
			found = counts != nil
			if status != "available" {
				t.Fatalf("status=%v", status)
			}
		} else if counts != nil {
			t.Fatal("cross-pool counts")
		} else if status != "unavailable" {
			t.Fatalf("unverified identity status=%v", status)
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

func TestMirrorLeaderCountsAvailability(t *testing.T) {
	for _, scenario := range []string{"command_failure", "parse_failure", "disabled"} {
		t.Run(scenario, func(t *testing.T) {
			override := map[string][]byte{
				"collect.rbd_mirroring":        []byte(`{"mode":"image","peers":[]}`),
				"collect.rbd_mirroring_status": []byte(`{"daemons":[]}`),
			}
			if scenario == "parse_failure" {
				override["collect.rbd_mirror_service_status"] = []byte(`invalid`)
			}
			if scenario == "disabled" {
				override["collect.rbd_mirroring"] = []byte(`{"mode":"disabled"}`)
			}
			var calls []executor.CommandSpec
			p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: override}}}
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			rows := p.collectStorageOptional(ctx, ClusterAccess{}, []poolWire{{Pool: 7, PoolName: "pool"}}, fsDumpWire{}, time.Now())
			found := false
			for _, row := range rows {
				if row.Kind != "rbd_mirroring" {
					continue
				}
				found = true
				payload := row.Payload.(map[string]any)
				want := "unavailable"
				if scenario == "disabled" {
					want = "disabled"
				}
				if payload["leader_counts_status"] != want || payload["leader_counts"] != nil {
					t.Fatalf("payload=%#v", payload)
				}
			}
			if !found {
				t.Fatal("missing inventory")
			}
			_, unavailable := trace.unavailable["rbd_mirroring"]
			if unavailable != (scenario != "disabled") {
				t.Fatalf("failure classification=%v", unavailable)
			}
			if scenario == "disabled" {
				for _, call := range calls {
					if call.ID == "collect.rbd_mirror_service_status" {
						t.Fatal("queried disabled pool")
					}
				}
			}
		})
	}
}
