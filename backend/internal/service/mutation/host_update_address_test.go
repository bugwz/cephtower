package mutation

import (
	"context"
	"reflect"
	"testing"
)

func TestHostAddressAndLabelsExecuteTogether(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`[{"hostname":"node1","addr":"10.0.0.2","labels":["new","keep"]}]`, true},
		{`[{"hostname":"node1","addr":"10.0.0.1","labels":["new"]}]`, false},
		{`[{"hostname":"node1","addr":"10.0.0.2","labels":["old"]}]`, false},
		{`[{"hostname":"node1","addr":null,"labels":["new"]}]`, false},
	} {
		service, _, id := newCephUserService(t)
		runner := &hostRemovalExecutor{raw: tc.raw}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.update", ResourceKey: "host/node1", Parameters: map[string]any{"address": "10.0.0.2", "labels_add": []any{"new"}, "labels_remove": []any{"old"}}})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
		want := [][]string{{"orch", "host", "set-addr", "node1", "10.0.0.2"}, {"orch", "host", "label", "add", "node1", "new"}, {"orch", "host", "label", "rm", "node1", "old"}, {"orch", "host", "ls", "--detail", "--format", "json"}}
		if len(runner.specs) != len(want) {
			t.Fatal(runner.specs)
		}
		for i, spec := range runner.specs {
			if !reflect.DeepEqual(spec.Args, want[i]) || spec.Mutating != (i < 3) {
				t.Fatal(spec)
			}
		}
	}
}

func TestHostAddressUpdateValidation(t *testing.T) {
	for _, p := range []map[string]any{
		{"address": ""}, {"address": "--force"}, {"address": " 10.0.0.2"},
		{"address": "10.0.0.2", "labels_add": []any{"duplicate", "duplicate"}},
		{"address": "10.0.0.2", "labels_remove": "wrong-type"},
	} {
		if _, err := build(Request{Action: "host.update", ResourceKey: "host/node1"}, p); err == nil {
			t.Fatal(p)
		}
	}
	if !hostUpdateMatches([]byte(`[{"hostname":"node1","addr":"10.0.0.2"}]`), "node1", map[string]any{"address": "10.0.0.2"}) {
		t.Fatal("address-only update requires unrelated label fields")
	}
}
