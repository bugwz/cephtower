package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestDeviceCapacitySerialization(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"0", "0"}, {"9007199254740993", "9007199254740993"}, {"18446744073709551615", "18446744073709551615"}, {"18446744073709551615.0", "18446744073709551615"},
		{"null", ""}, {"-1", ""}, {"1.5", ""}, {"18446744073709551616", ""}, {`"01"`, ""}, {"1e3", ""},
	} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.device": []byte(`[{"name":"node1","devices":[{"path":"/dev/sda","sys_api":{"size":` + tc.raw + `}}]}]`)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "inventory")
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatalf("rows=%v", rows)
		}
		payload := rows[0].Payload.(cephdomain.Device)
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]any
		if err := json.Unmarshal(encoded, &data); err != nil {
			t.Fatal(err)
		}
		if tc.want == "" {
			if _, ok := data["size_bytes"]; ok {
				t.Fatalf("invalid %s accepted: %s", tc.raw, encoded)
			}
		} else if data["size_bytes"] != tc.want {
			t.Fatalf("%s: %s", tc.raw, encoded)
		}
	}
}
