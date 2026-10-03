package clusterinspect

import (
	"context"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type perfFixtureExecutor struct {
	schema, dump string
	specs        []executor.CommandSpec
}

func (e *perfFixtureExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	output := e.schema
	if spec.ID == "daemon.perf.dump" {
		output = e.dump
	}
	return executor.CommandResult{Stdout: []byte(output)}, nil
}

func TestDaemonPerfSnapshotPreservesNativeValues(t *testing.T) {
	s, _, id := testInspection(t)
	e := &perfFixtureExecutor{schema: `{"osd":{"bytes":{"description":"bytes","units":"bytes","type":10,"priority":5},"latency":{"value_type":"real-integer-pair"},"missing":{}}}`, dump: `{"osd":{"bytes":18446744073709551615,"latency":{"avgcount":9007199254740993,"sum":1.5}}}`}
	s.executor = e
	result, err := s.DaemonPerf(context.Background(), id, "osd.1")
	if err != nil {
		t.Fatal(err)
	}
	rows := result["items"].([]map[string]any)
	if len(rows) != 3 || rows[0]["raw_value"] != "18446744073709551615" || rows[1]["raw_value"] != `{"avgcount":9007199254740993,"sum":1.5}` || rows[2]["raw_value"] != nil || result["daemon_name"] != "osd.1" {
		t.Fatalf("%+v", result)
	}
	for i, section := range []string{"schema", "dump"} {
		if e.specs[i].Mutating || !reflect.DeepEqual(e.specs[i].Args, []string{"tell", "osd.1", "perf", section, "--format", "json"}) {
			t.Fatalf("%+v", e.specs)
		}
	}
}

func TestDaemonPerfRejectsInvalidResponsesAndNames(t *testing.T) {
	s, _, id := testInspection(t)
	for _, schema := range []string{"null", "[]", "{} {}", `{"osd":null}`, `{"osd":{"x":null}}`, `{"osd":{"x":{"type":"counter"}}}`} {
		s.executor = &perfFixtureExecutor{schema: schema, dump: `{}`}
		if _, err := s.DaemonPerf(context.Background(), id, "osd.1"); err == nil {
			t.Fatalf("accepted %s", schema)
		}
	}
	e := &perfFixtureExecutor{schema: `{}`, dump: `{}`}
	s.executor = e
	for _, name := range []string{"--help", "osd.", "osd.*", "grafana.node", "osd.1;stop"} {
		if _, err := s.DaemonPerf(context.Background(), id, name); err == nil {
			t.Fatalf("accepted %s", name)
		}
	}
	if len(e.specs) != 0 {
		t.Fatal("invalid identity executed")
	}
	result, err := s.DaemonPerf(context.Background(), id, "mgr.a")
	if err != nil || len(result["items"].([]map[string]any)) != 0 {
		t.Fatalf("empty snapshot failed: %v", err)
	}
}
