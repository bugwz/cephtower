package mutation

import (
	"encoding/json"
	"testing"
)

func TestServicePlacementNativeConstraints(t *testing.T) {
	for _, raw := range []string{`{}`, `{"count":2}`, `{"count_per_host":2,"label":"rgw"}`, `{"count_per_host":1,"hosts":["node-a"]}`, `{"count_per_host":3,"host_pattern":"node-*"}`} {
		var p map[string]any
		json.Unmarshal([]byte(raw), &p)
		cmd, err := build(Request{Action: "service.create"}, map[string]any{"service_type": "rgw", "service_id": "a", "placement": p})
		if err != nil {
			t.Fatal(raw, err)
		}
		var out struct {
			Placement map[string]any `json:"placement"`
		}
		json.Unmarshal(cmd.stdin, &out)
		if len(out.Placement) != len(p) {
			t.Fatalf("placement fields dropped: %s", cmd.stdin)
		}
	}
	for _, raw := range []string{`{"count":0}`, `{"count":1.2}`, `{"count":null}`, `{"count":"2"}`, `{"count_per_host":0,"label":"x"}`, `{"count_per_host":1}`, `{"count":1,"count_per_host":2,"label":"x"}`, `{"hosts":["a"],"label":"x"}`, `{"hosts":["a"],"host_pattern":"*"}`, `{"count_per_host":1,"hosts":["a=name"]}`, `{"count_per_host":1,"hosts":["a:1.2.3.4"]}`, `{"hosts":[""]}`, `{"typo":1}`} {
		var p map[string]any
		json.Unmarshal([]byte(raw), &p)
		if err := validateServicePlacement(p); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
