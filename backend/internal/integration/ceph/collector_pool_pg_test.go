package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestCollectPoolPGStates(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		want           map[int64]map[string]uint64
	}{
		{"mixed", `{"pg_stats":[{"pgid":"7.a","state":"active+clean"},{"pgid":"7.b","state":"active+clean"},{"pgid":"7.c","state":"active+degraded"},{"pgid":"2.0","state":"down"}]}`, map[int64]map[string]uint64{7: {"active+clean": 2, "active+degraded": 1}, 2: {"down": 1}}},
		{"empty", `{"pg_stats":[]}`, map[int64]map[string]uint64{}},
		{"missing", `{}`, nil},
		{"null", `{"pg_stats":null}`, nil},
		{"bad id", `{"pg_stats":[{"pgid":"oops","state":"down"}]}`, nil},
		{"bad seed", `{"pg_stats":[{"pgid":"7.xyz","state":"down"}]}`, nil},
		{"missing state", `{"pg_stats":[{"pgid":"7.0"}]}`, nil},
		{"duplicate", `{"pg_stats":[{"pgid":"7.0","state":"down"},{"pgid":"7.0","state":"down"}]}`, nil},
		{"partial", `{"pg_stats":[{"pgid":"7.0","state":"down"},null]}`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []executor.CommandSpec
			p := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.pool_pg_states": []byte(tc.response)}}, calls: &calls}}
			if got := p.collectPoolPGStates(context.Background(), ClusterAccess{}); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v want %#v", got, tc.want)
			}
			if len(calls) != 1 || !reflect.DeepEqual(calls[0].Args, []string{"pg", "dump", "pgs_brief", "--format", "json"}) {
				t.Fatalf("commands = %#v", calls)
			}
		})
	}
	provider := NativeProvider{Executor: fixtureExecutor{t}}
	if got := provider.collectPoolPGStates(context.Background(), ClusterAccess{}); got != nil {
		t.Fatalf("unavailable command = %#v", got)
	}
}

func TestPoolPGStatesReachInventoryPayload(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool":           []byte(`[{"pool":7,"pool_name":"pg-pool","type":3,"erasure_code_profile":"archive-ec"}]`),
		"collect.pool_pg_states": []byte(`{"pg_stats":[{"pgid":"7.0","state":"active+degraded"},{"pgid":"8.0","state":"down"}]}`),
	}}}
	rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Kind != "pool" {
			continue
		}
		payload := row.Payload.(cephdomain.Pool)
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded["pg_status"], map[string]any{"active+degraded": float64(1)}) {
			t.Fatalf("payload = %s", data)
		}
		if decoded["erasure_code_profile"] != "archive-ec" {
			t.Fatalf("native erasure profile missing from inventory: %s", data)
		}
		return
	}
	t.Fatal("pool observation missing")
}

func TestPoolProfileMissingStaysUnknown(t *testing.T) {
	var wire poolWire
	if err := json.Unmarshal([]byte(`{"pool":1,"pool_name":"unknown","type":3}`), &wire); err != nil {
		t.Fatal(err)
	}
	if wire.ErasureCodeProfile != nil {
		t.Fatalf("invented profile: %v", wire.ErasureCodeProfile)
	}
	if err := json.Unmarshal([]byte(`{"pool":1,"pool_name":"bad","erasure_code_profile":42}`), &wire); err == nil {
		t.Fatal("numeric profile accepted")
	}
}
