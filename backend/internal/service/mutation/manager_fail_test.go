package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"reflect"
	"testing"
)

func TestManagerFailUsesSelectedTarget(t *testing.T) {
	for _, name := range []string{"a", "mgr.a", "host.manager-1", "fail"} {
		spec, err := build(Request{Action: "manager.fail", ResourceKey: "manager/" + name + "/fail"}, map[string]any{"name": name})
		if err != nil {
			t.Fatal(err)
		}
		if spec.binary != executor.BinaryCeph || !reflect.DeepEqual(spec.args, []string{"mgr", "fail", name}) || !reflect.DeepEqual(spec.check, []string{"mgr", "dump", "--format", "json"}) {
			t.Fatalf("wrong command: %+v", spec)
		}
	}
}

func TestManagerFailRejectsAmbiguousTargets(t *testing.T) {
	for _, key := range []string{"", "manager/a", "manager//fail", "/manager/a/fail", "manager/a/fail/", "manager/a/other", "other/a/fail", "manager/a/b/fail", "manager/ a/fail", "manager/a /fail", "manager/--help/fail", "manager/a\nb/fail", "manager/../fail"} {
		if _, err := build(Request{Action: "manager.fail", ResourceKey: key}, nil); err == nil {
			t.Fatalf("accepted %q", key)
		}
	}
	for _, name := range []any{"other", 1, nil, []any{"a"}, map[string]any{"name": "a"}} {
		if _, err := build(Request{Action: "manager.fail", ResourceKey: "manager/a/fail"}, map[string]any{"name": name}); err == nil {
			t.Fatalf("accepted mismatched name %#v", name)
		}
	}
}
