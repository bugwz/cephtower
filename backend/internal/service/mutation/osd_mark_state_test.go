package mutation

import (
	"context"
	"testing"
)

func TestOSDMarkStateReadback(t *testing.T) {
	for _, tc := range []struct {
		action, raw string
		valid       bool
	}{
		{"in", `{"osds":[{"osd":0,"in":1}]}`, true},
		{"out", `{"osds":[{"osd":0,"in":0}]}`, true},
		{"down", `{"osds":[{"osd":0,"up":0}]}`, true},
		{"down", `{"osds":[{"osd":0,"up":1}]}`, false},
		{"in", `{"osds":[{"osd":0,"in":0}]}`, false},
		{"out", `{"osds":[{"osd":0,"in":1}]}`, false},
		{"out", `{"osds":[{"osd":0,"in":null}]}`, false},
		{"down", `{"osds":[{"osd":0}]}`, false},
		{"out", `{"osds":[{"osd":1,"in":0}]}`, false},
		{"out", `{"osds":[{"in":0}]}`, false},
		{"out", `{"osds":[{"osd":0,"in":0},{"osd":0,"in":0}]}`, false},
		{"out", `{"osds":[]}`, false},
		{"in", `{"osds":[{"osd":0,"in":2}]}`, false},
	} {
		service, _, id := newCephUserService(t)
		runner := &osdSafetyExecutor{output: tc.raw}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.action", ResourceKey: "osd/0/action", Parameters: map[string]any{"action": tc.action}})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
		if len(runner.specs) != 2 || runner.specs[1].Mutating {
			t.Fatal(runner.specs)
		}
	}
}
