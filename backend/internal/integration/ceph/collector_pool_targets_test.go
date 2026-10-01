package ceph

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPoolPGTargetsReachInventory(t *testing.T) {
	for _, targets := range []bool{false, true} {
		response := `[{"pool":7,"pool_name":"targets","type":1,"pg_num":32,"pg_placement_num":16}]`
		if targets {
			response = `[{"pool":7,"pool_name":"targets","type":1,"pg_num":32,"pg_placement_num":16,"pg_num_target":128,"pg_placement_num_target":64}]`
		}
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.pool": []byte(response)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "pool" {
				continue
			}
			found = true
			data, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded["pg_num"] != float64(32) || decoded["pgp_num"] != float64(16) {
				t.Fatalf("current counts: %s", data)
			}
			if targets {
				if decoded["pg_num_target"] != float64(128) || decoded["pgp_num_target"] != float64(64) {
					t.Fatalf("target counts: %s", data)
				}
			} else if decoded["pg_num_target"] != nil || decoded["pgp_num_target"] != nil {
				t.Fatalf("invented targets: %s", data)
			}
		}
		if !found {
			t.Fatal("missing pool")
		}
	}
}
