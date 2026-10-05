package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestRGWOwnerStorageCountersRemainExact(t *testing.T) {
	for _, kind := range []string{"rgw_user", "rgw_account"} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect." + kind:             []byte(`["owner"]`),
			"collect." + kind + "_detail": []byte(`{"id":"owner","user_id":"owner","name":"owner"}`),
			"collect." + kind + "_stats":  []byte(`{"stats":{"size":18446744073709551615,"size_actual":9007199254740993,"size_utilized":0,"num_objects":9007199254740995,"size_kb":1},"last_synced":"raw-time","last_stats_sync":"raw-time"}`),
		}}}
		found := false
		for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
			if row.Kind != kind {
				continue
			}
			found = true
			encoded, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(encoded, &payload); err != nil {
				t.Fatal(err)
			}
			stats := payload["storage_stats"].(map[string]any)
			counters := stats["stats"].(map[string]any)
			for field, want := range map[string]string{"size": "18446744073709551615", "size_actual": "9007199254740993", "size_utilized": "0", "num_objects": "9007199254740995", "size_kb": "1"} {
				if counters[field] != want {
					t.Fatalf("%s/%s lost precision: %v", kind, field, counters[field])
				}
			}
			if stats["last_synced"] != "raw-time" || stats["last_stats_sync"] != "raw-time" {
				t.Fatal("timestamps changed")
			}
			if _, exists := counters["size_kb_actual"]; exists {
				t.Fatal("invented missing counter")
			}
		}
		if !found {
			t.Fatalf("missing %s", kind)
		}
	}
}
