package clusterinspect

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
)

func TestRGWTopologyCounts(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		kind := spec.Args[0]
		if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating || !reflect.DeepEqual(spec.Args, []string{kind, "list", "--format", "json"}) {
			t.Fatal("incorrect native list")
		}
		return executor.CommandResult{Stdout: []byte(`{"` + kind + `s":["a","b"]}`)}, nil
	}
	result, err := s.RGWTopologyCounts(context.Background(), id)
	if err != nil || result["realm_count"] != 2 || result["zonegroup_count"] != 2 || result["zone_count"] != 2 || len(runner.specs) != 3 {
		t.Fatal("invalid counts", result, err)
	}
	for _, raw := range []string{`null`, `{}`, `{"realms":null}`, `{"realms":["a","a"]}`, `{"realms":[""]}`, `{"realms":[1]}`, `{"realms":[]} {}`} {
		runner.run = nil
		runner.output = raw
		if result, err = s.RGWTopologyCounts(context.Background(), id); err == nil || result != nil {
			t.Fatal("invalid count accepted", raw)
		}
	}
	for _, failKind := range []string{"realm", "zonegroup", "zone"} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			kind := spec.Args[0]
			exit := 0
			if kind == failKind {
				exit = 2
			}
			return executor.CommandResult{Stdout: []byte(`{"` + kind + `s":[]}`), ExitCode: exit}, nil
		}
		if result, err = s.RGWTopologyCounts(context.Background(), id); err == nil || result != nil {
			t.Fatal("partial counts returned")
		}
	}
}
