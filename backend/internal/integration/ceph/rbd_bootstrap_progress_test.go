package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestMirrorBootstrapPercent(t *testing.T) {
	for _, percent := range []string{"0", "1", "99", "100"} {
		row := map[string]any{"state": "up+syncing", "description": "bootstrapping, IMAGE_SYNC/COPY_IMAGE " + percent + "%"}
		if got := mirrorBootstrapPercent(row); got != percent {
			t.Fatalf("percent=%q, want %q", got, percent)
		}
		for _, state := range []any{"down+syncing", "up+starting_replay", "up+replaying", "up+stopped", nil} {
			row["state"] = state
			if got := mirrorBootstrapPercent(row); got != "" {
				t.Fatalf("accepted state %v: %q", state, got)
			}
		}
	}
	for _, description := range []any{nil, 0, "bootstrapping, IMAGE_SYNC/COPY_IMAGE", "bootstrapping, IMAGE_COPY/COPY_OBJECT 50%", "bootstrapping, IMAGE_SYNC/COPY_IMAGE -1%", "bootstrapping, IMAGE_SYNC/COPY_IMAGE 101%", "bootstrapping, IMAGE_SYNC/COPY_IMAGE 50.5%", "bootstrapping, IMAGE_SYNC/COPY_IMAGE 01%", "bootstrapping, IMAGE_SYNC/COPY_IMAGE 50%\n", "bootstrapping, IMAGE_SYNC/FLUSH_SYNC_POINT"} {
		if got := mirrorBootstrapPercent(map[string]any{"state": "up+syncing", "description": description}); got != "" {
			t.Fatalf("accepted description %#v: %q", description, got)
		}
	}
}

func TestCollectMirrorBootstrapPercent(t *testing.T) {
	image := map[string]any{"name": "image", "state": "up+syncing", "description": "bootstrapping, IMAGE_SYNC/COPY_IMAGE 0%", "peer_sites": []any{map[string]any{"state": "up+syncing", "description": "bootstrapping, IMAGE_SYNC/COPY_IMAGE 100%"}}}
	raw, _ := json.Marshal(map[string]any{"images": []any{image}})
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_mirroring": []byte(`{"mode":"image","peers":[]}`), "collect.rbd_mirroring_status": raw,
	}}}
	for _, row := range provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now()) {
		if row.Kind != "rbd_mirroring" {
			continue
		}
		got := row.Payload.(map[string]any)["images"].([]any)[0].(map[string]any)
		if got["bootstrap_percent"] != "0" || got["description"] != image["description"] {
			t.Fatalf("local bootstrap: %#v", got)
		}
		peer := got["peer_sites"].([]any)[0].(map[string]any)
		if peer["bootstrap_percent"] != "100" {
			t.Fatalf("remote bootstrap: %#v", peer)
		}
		if got["replay_metrics"] != nil || peer["replay_metrics"] != nil {
			t.Fatal("bootstrap invented replay metrics")
		}
		return
	}
	t.Fatal("missing mirror inventory")
}
