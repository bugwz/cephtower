package ceph

import (
	"encoding/json"
	"testing"
)

func TestPoolCacheTierDetailPreservesCounters(t *testing.T) {
	var wire poolWire
	if err := json.Unmarshal([]byte(`{"pool":7,"pool_name":"cache","type":1,"tiers":[],"cache_mode":"writeback","cache_min_evict_age":0,"cache_min_flush_age":10,"target_max_bytes":18446744073709551615,"target_max_objects":9007199254740993}`), &wire); err != nil {
		t.Fatal(err)
	}
	detail := poolRawDetail(wire.Raw, "replicated")
	encoded, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatal(err)
	}
	for key, expected := range map[string]string{"cache_min_evict_age": "0", "cache_min_flush_age": "10", "target_max_bytes": "18446744073709551615", "target_max_objects": "9007199254740993"} {
		if detail[key] != expected {
			t.Fatalf("%s lost precision: %#v", key, detail[key])
		}
		if response[key] != expected {
			t.Fatalf("API serialization lost %s: %s", key, encoded)
		}
		if _, ok := wire.Raw[key].(json.Number); !ok {
			t.Fatal("mutated native input")
		}
	}
	if detail["cache_mode"] != "writeback" {
		t.Fatal(detail)
	}
	if _, exists := poolRawDetail(map[string]any{"pool": json.Number("1")}, "replicated")["target_max_bytes"]; exists {
		t.Fatal("invented missing counter")
	}
}
