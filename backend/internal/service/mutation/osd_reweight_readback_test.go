package mutation

import (
	"context"
	"encoding/json"
	"testing"
)

func TestOSDReweightNativeQuantization(t *testing.T) {
	for _, tc := range []struct {
		weight, raw string
		valid       bool
	}{
		{"0", `{"osds":[{"osd":0,"weight":0}]}`, true},
		{"1", `{"osds":[{"osd":0,"weight":1}]}`, true},
		{"0.5", `{"osds":[{"osd":0,"weight":0.5}]}`, true},
		{"0.1", `{"osds":[{"osd":0,"weight":0.0999908447265625}]}`, true},
		{"0.1", `{"osds":[{"osd":0,"weight":0.1}]}`, false},
		{"0.000001", `{"osds":[{"osd":0,"weight":0}]}`, true},
		{"0", `{"osds":[{"osd":0}]}`, false},
		{"0", `{"osds":[{"osd":0,"weight":null}]}`, false},
		{"0", `{"osds":[{"osd":1,"weight":0}]}`, false},
		{"0", `{"osds":[{"osd":0,"weight":0},{"osd":0,"weight":0}]}`, false},
	} {
		service, _, id := newCephUserService(t)
		service.executor = &osdSafetyExecutor{output: tc.raw}
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.action", ResourceKey: "osd/0/action", Parameters: map[string]any{"action": "reweight", "weight": json.Number(tc.weight)}})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
	}
}
