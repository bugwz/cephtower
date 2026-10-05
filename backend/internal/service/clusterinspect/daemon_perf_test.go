package clusterinspect

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type perfFixtureExecutor struct {
	schema, dump string
	specs        []executor.CommandSpec
	failAt       string
}

func TestDaemonPerfRedactsCompositeValues(t *testing.T) {
	s, _, id := testInspection(t)
	s.executor = &perfFixtureExecutor{schema: `{"osd":{"value":{}}}`, dump: `{"osd":{"value":{"avgcount":18446744073709551615,"sum":1.25,"password":"fixture-password","extra":[{"to\u006ben":"fixture-token"}]}}}`}
	result, err := s.DaemonPerf(context.Background(), id, "osd.1")
	if err != nil {
		t.Fatal(err)
	}
	value := result["items"].([]map[string]any)[0]["raw_value"].(string)
	if strings.Contains(value, "fixture-password") || strings.Contains(value, "fixture-token") || !strings.Contains(value, "[REDACTED]") || !strings.Contains(value, `"avgcount":18446744073709551615`) || !strings.Contains(value, `"sum":1.25`) {
		t.Fatal("performance value leaked credentials or lost numeric precision")
	}
}

func (e *perfFixtureExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	output := e.schema
	if spec.ID == "daemon.perf.dump" {
		output = e.dump
	}
	exit := 0
	if spec.ID == e.failAt {
		exit = 2
	}
	return executor.CommandResult{Stdout: []byte(output), ExitCode: exit, Stderr: []byte("private-diagnostic")}, nil
}

func TestDaemonPerfRejectsFailedCommandsWithValidJSON(t *testing.T) {
	for _, stage := range []string{"daemon.perf.schema", "daemon.perf.dump"} {
		s, _, id := testInspection(t)
		e := &perfFixtureExecutor{schema: `{"osd":{"x":{}}}`, dump: `{"osd":{"x":123}}`, failAt: stage}
		s.executor = e
		result, err := s.DaemonPerf(context.Background(), id, "osd.1")
		if result != nil || err == nil || strings.Contains(err.Error(), "private-diagnostic") {
			t.Fatal("failed command became a snapshot or leaked diagnostics")
		}
		want := 1
		if stage == "daemon.perf.dump" {
			want = 2
		}
		if len(e.specs) != want {
			t.Fatal("continued after failed command")
		}
	}
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
	for _, schema := range []string{"null", "[]", "{} {}", `{"osd":null}`, `{"osd":{"x":null}}`, `{"osd":{"x":{"type":"counter"}}}`, `{" ":{"x":{}}}`, `{"osd":{"":{}}}`, `{"osd":{"a.b":{}},"osd.a":{"b":{}}}`} {
		s.executor = &perfFixtureExecutor{schema: schema, dump: `{}`}
		if _, err := s.DaemonPerf(context.Background(), id, "osd.1"); err == nil {
			t.Fatalf("accepted %s", schema)
		}
	}
	for _, dump := range []string{"null", "[]", "{} {}", `{"osd":null}`, `{"osd":[]}`, `{" ":{}}`} {
		s.executor = &perfFixtureExecutor{schema: `{"osd":{"x":{}}}`, dump: dump}
		if result, err := s.DaemonPerf(context.Background(), id, "osd.1"); err == nil || result != nil {
			t.Fatalf("accepted malformed dump %s: %+v", dump, result)
		}
	}
	e := &perfFixtureExecutor{schema: `{}`, dump: `{}`}
	s.executor = e
	for _, name := range []string{"--help", "osd.", "osd.*", "grafana.node", "osd.1;stop", "rgw.a", "rgw.123", "rbd-mirror.a"} {
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
