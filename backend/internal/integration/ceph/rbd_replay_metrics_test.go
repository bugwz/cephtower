package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMirrorReplayMetrics(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want map[string]string
	}{
		{`{"replay_state":"idle","bytes_per_second":0}`, map[string]string{"replay_state": "idle", "bytes_per_second": "0"}},
		{`{"replay_state":"syncing"}`, map[string]string{"replay_state": "syncing"}},
		{`{"replay_state":"future"}`, map[string]string{}},
		{`{"replay_state":false}`, map[string]string{}},
		{`{"bytes_per_second":0,"syncing_percent":100,"seconds_until_synced":18446744073709551615,"entries_behind_primary":9007199254740993}`, map[string]string{"bytes_per_second": "0", "syncing_percent": "100", "seconds_until_synced": "18446744073709551615", "entries_behind_primary": "9007199254740993"}},
		{`{"bytes_per_second":1.25e3,"syncing_percent":0.5}`, map[string]string{"bytes_per_second": "1.25e3", "syncing_percent": "0.5"}},
		{`{"bytes_per_second":-1,"syncing_percent":100.000000000000000001,"seconds_until_synced":1.5,"entries_behind_primary":18446744073709551616}`, map[string]string{}},
		{`{"bytes_per_second":1e999,"syncing_percent":1e-999,"seconds_until_synced":"0","entries_behind_primary":null}`, map[string]string{}},
		{`{"bytes_per_second":true,"syncing_percent":[]}`, map[string]string{}},
	} {
		row := map[string]any{"state": "up+replaying", "description": "replaying, " + tc.raw}
		if got := mirrorReplayMetrics(row); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: %#v != %#v", tc.raw, got, tc.want)
		}
		row["state"] = "down+replaying"
		if got := mirrorReplayMetrics(row); got != nil {
			t.Fatalf("down state exposed metrics: %#v", got)
		}
	}
	for _, raw := range []string{"replaying", "replaying, []", "replaying, {} trailing", "replaying, {"} {
		if got := mirrorReplayMetrics(map[string]any{"state": "up+replaying", "description": raw}); len(got) != 0 {
			t.Fatalf("invalid description: %#v", got)
		}
	}
}

func TestCollectMirrorReplayMetrics(t *testing.T) {
	image := map[string]any{"name": "image", "state": "up+replaying", "description": `replaying, {"bytes_per_second":0,"replay_state":"idle"}`, "peer_sites": []any{map[string]any{"state": "up+replaying", "description": `replaying, {"entries_behind_primary":9007199254740993,"replay_state":"syncing"}`}}}
	raw, _ := json.Marshal(map[string]any{"images": []any{image}})
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_mirroring": []byte(`{"mode":"image","peers":[]}`), "collect.rbd_mirroring_status": raw,
	}}}
	for _, row := range provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now()) {
		if row.Kind != "rbd_mirroring" {
			continue
		}
		got := row.Payload.(map[string]any)["images"].([]any)[0].(map[string]any)
		if got["description"] != image["description"] || got["replay_metrics"].(map[string]string)["bytes_per_second"] != "0" {
			t.Fatalf("local metrics: %#v", got)
		}
		peer := got["peer_sites"].([]any)[0].(map[string]any)
		if got["replay_metrics"].(map[string]string)["replay_state"] != "idle" || peer["replay_metrics"].(map[string]string)["replay_state"] != "syncing" {
			t.Fatalf("mixed local and peer replay states: %#v", got)
		}
		if peer["replay_metrics"].(map[string]string)["entries_behind_primary"] != "9007199254740993" {
			t.Fatalf("remote metrics: %#v", peer)
		}
		return
	}
	t.Fatal("missing mirror inventory")
}
