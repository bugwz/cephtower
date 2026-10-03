package clusterinspect

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestErasureCodeInfo(t *testing.T) {
	s, runner, id := testInspection(t)
	mgr := `{"active_name":"node.a"}`
	config := `[{"name":"osd_erasure_code_plugins","value":"jerasure isa lrc isa custom"},{"name":"erasure_code_dir","value":"/usr/lib/ceph/erasure-code"},{"name":"unrelated_secret","value":"not-exported"}]`
	fail := false
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if fail {
			return executor.CommandResult{}, errors.New("unavailable")
		}
		value := config
		if spec.ID == "erasure_code.manager" {
			value = mgr
		}
		return executor.CommandResult{Stdout: []byte(value)}, nil
	}
	result, err := s.ErasureCodeInfo(context.Background(), id)
	if err != nil || result.Manager != "mgr.node.a" || result.Directory != "/usr/lib/ceph/erasure-code" || result.ObservedAt.IsZero() || !reflect.DeepEqual(result.Plugins, []string{"jerasure", "isa", "lrc", "custom"}) {
		t.Fatal(result, err)
	}
	for i, args := range [][]string{{"mgr", "dump", "--format", "json"}, {"config", "show-with-defaults", "mgr.node.a", "--format", "json"}} {
		if runner.specs[i].Mutating || !reflect.DeepEqual(runner.specs[i].Args, args) {
			t.Fatal(runner.specs[i])
		}
	}
	for _, output := range []string{`null`, `{}`, `[]`, `[{"name":"erasure_code_dir","value":"/lib"}]`, `[{"name":"osd_erasure_code_plugins","value":" "},{"name":"erasure_code_dir","value":"/lib"}]`, `[{"name":"osd_erasure_code_plugins","value":"isa"},{"name":"erasure_code_dir","value":null}]`, `[{"name":"osd_erasure_code_plugins","value":"isa"},{"name":"erasure_code_dir","value":1}]`, `[{"name":"osd_erasure_code_plugins","value":"isa"},{"name":"erasure_code_dir","value":"/lib"},{"name":"erasure_code_dir","value":"/other"}]`, `[] {}`} {
		config = output
		if _, err := s.ErasureCodeInfo(context.Background(), id); err == nil {
			t.Fatal("accepted", output)
		}
	}
	for _, output := range []string{`null`, `{}`, `{"active_name":""}`, `{"active_name":"--help"}`, `{"active_name":"a b"}`} {
		mgr = output
		runner.specs = nil
		if _, err := s.ErasureCodeInfo(context.Background(), id); err == nil || len(runner.specs) != 1 {
			t.Fatal(output, err, runner.specs)
		}
	}
	fail = true
	if _, err := s.ErasureCodeInfo(context.Background(), id); err == nil {
		t.Fatal("failure hidden")
	}
}
