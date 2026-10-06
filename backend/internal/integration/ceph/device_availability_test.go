package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestDeviceAvailabilityUnknown(t *testing.T) {
	for _, tc := range []struct {
		field string
		want  any
	}{
		{`,"available":true`, true}, {`,"available":false`, false}, {"", nil}, {`,"available":null`, nil}, {`,"available":"false"`, nil}, {`,"available":1`, nil},
	} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.device": []byte(`[{"hostname":"node1","path":"/dev/sda"` + tc.field + `}]`)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "inventory")
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatal(rows)
		}
		encoded, err := json.Marshal(rows[0].Payload.(cephdomain.Device))
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]any
		if err := json.Unmarshal(encoded, &data); err != nil {
			t.Fatal(err)
		}
		if value, present := data["available"]; !present || value != tc.want {
			t.Fatalf("%s: %s", tc.field, encoded)
		}
	}
}
