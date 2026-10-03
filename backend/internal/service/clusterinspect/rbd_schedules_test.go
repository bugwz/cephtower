package clusterinspect

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestLiveRBDMirrorSchedules(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `[{"pool":"-","namespace":"-","image":"-","items":[{"interval":"1h","start_time":"01:00:00+08:00"}],"unrelated":"not-exported"},{"pool":"images","namespace":"","image":"vm","items":[{"interval":"2h","start_time":""}]}]`
	result, err := s.RBDMirrorSchedules(context.Background(), id)
	if err != nil || result.ObservedAt.IsZero() || !strings.Contains(string(result.Schedules), `"01:00:00+08:00"`) || strings.Contains(string(result.Schedules), "not-exported") {
		t.Fatal(result, err)
	}
	if len(runner.specs) != 1 || runner.specs[0].Binary != executor.BinaryRBD || runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"mirror", "snapshot", "schedule", "list", "--recursive", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	runner.output = `[]`
	if result, err := s.RBDMirrorSchedules(context.Background(), id); err != nil || string(result.Schedules) != "[]" {
		t.Fatal(result, err)
	}
	for _, output := range []string{`null`, `{}`, `[] {}`, `[{}]`, `[{"pool":"p","image":"vm","items":[]}]`, `[{"pool":"-","namespace":"team","image":"-","items":[]}]`, `[{"pool":"-","namespace":"-","image":"-","items":[{"interval":"0h"}]}]`, `[{"pool":"-","namespace":"-","image":"-","items":[]},{"pool":"-","namespace":"-","image":"-","items":[]}]`} {
		runner.output = output
		if _, err := s.RBDMirrorSchedules(context.Background(), id); err == nil {
			t.Fatal("invalid list accepted", output)
		}
	}
	runner.fail = true
	if _, err := s.RBDMirrorSchedules(context.Background(), id); err == nil {
		t.Fatal("failure became empty list")
	}
}
