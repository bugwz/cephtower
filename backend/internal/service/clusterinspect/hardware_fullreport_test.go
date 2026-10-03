package clusterinspect

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestHardwareFullReport(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"node1":{"host":"node1","sn":"00123","status":{"memory":{"system":{"dimm":{"capacity_bytes":18446744073709551615,"password":"fixture-secret"}}}},"token":"fixture-token"}}`
	result, err := s.Hardware(context.Background(), id, "node1", "fullreport")
	if err != nil {
		t.Fatal(err)
	}
	report := result["report"].(string)
	if result["host"] != "node1" || result["category"] != "fullreport" || result["serial_number"] != "00123" || !strings.Contains(report, "18446744073709551615") || strings.Contains(report, "fixture-secret") || strings.Contains(report, "fixture-token") {
		t.Fatalf("invalid report: %#v", result)
	}
	if runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "hardware", "status", "--hostname", "node1", "--category", "fullreport", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	before := len(runner.specs)
	if _, err := s.Hardware(context.Background(), id, "", "fullreport"); err == nil || len(runner.specs) != before {
		t.Fatal("hostless full report query executed")
	}
	for _, output := range []string{`null`, `[]`, `{}`, `{"other":{"host":"node1"}}`, `{"node1":{"host":"node1"},"other":{}}`, `{"node1":null}`, `{"node1":[]}`, `{"node1":{}}`, `{"node1":{"host":"other"}}`, `{"node1":{"host":"node1","sn":123}}`, `{"node1":{"host":"node1"}} {}`} {
		runner.output = output
		if result, err := s.Hardware(context.Background(), id, "node1", "fullreport"); err == nil || result != nil {
			t.Fatalf("accepted invalid report %s", output)
		}
	}
	for _, output := range []string{`{"node1":{"host":"node1"}}`, `{"node1":{"host":"node1","sn":null}}`} {
		runner.output = output
		result, err := s.Hardware(context.Background(), id, "node1", "fullreport")
		if err != nil || result["serial_number"] != nil {
			t.Fatalf("missing serial fabricated: %#v %v", result, err)
		}
	}
	runner.fail = true
	if _, err := s.Hardware(context.Background(), id, "node1", "fullreport"); err == nil {
		t.Fatal("command failure hidden")
	}
}
