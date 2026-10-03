package clusterinspect

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestHardwareCriticalsPreservesScopeAndNativeSeverity(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"node1":{"sys":{"memory":{"part":{"description":"DIMM A","status":{"health":"Warning","state":"Enabled"}}},"power":{"part":{"name":"PSU A","status":{"health":"Critical"},"counter":18446744073709551615,"password":"fixture-secret"}}}},"node2":{"sys":{"power":{"part":{"id":"PSU B"}}}}}`
	result, err := s.Hardware(context.Background(), id, "", "criticals")
	if err != nil {
		t.Fatal(err)
	}
	rows := result["items"].([]map[string]any)
	if len(rows) != 3 || rows[0]["health"] != "Warning" || rows[0]["name"] != "DIMM A" || rows[1]["health"] != "Critical" || rows[2]["health"] != nil || rows[2]["name"] != "PSU B" || rows[0]["id"] == rows[1]["id"] || rows[1]["id"] == rows[2]["id"] {
		t.Fatalf("%#v", result)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "fixture-secret") || !strings.Contains(rows[1]["details"].(string), "18446744073709551615") {
		t.Fatal("redaction or precision failure")
	}
	if runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "hardware", "status", "--category", "criticals", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	if _, err := s.Hardware(context.Background(), id, "node1", "criticals"); err == nil {
		t.Fatal("accepted other host in scoped report")
	}
	for _, output := range []string{`null`, `[]`, `{"node1":null}`, `{"node1":{"s":null}}`, `{"node1":{"s":{"memory":null}}}`, `{"node1":{"s":{"memory":{"p":null}}}}`, `{"node1":{"s":{"memory":{"p":{"status":{"health":false}}}}}}`, `{} {}`} {
		runner.output = output
		if _, err := s.Hardware(context.Background(), id, "node1", "criticals"); err == nil {
			t.Fatalf("accepted invalid report %s", output)
		}
	}
	for _, output := range []string{`{}`, `{"node1":{}}`, `{"node1":{"sys":{"memory":{}}}}`} {
		runner.output = output
		result, err := s.Hardware(context.Background(), id, "node1", "criticals")
		if err != nil || len(result["items"].([]map[string]any)) != 0 {
			t.Fatalf("%#v %v", result, err)
		}
	}
	last := runner.specs[len(runner.specs)-1]
	if !reflect.DeepEqual(last.Args, []string{"orch", "hardware", "status", "--hostname", "node1", "--category", "criticals", "--format", "json"}) {
		t.Fatal(last)
	}
	runner.fail = true
	if _, err := s.Hardware(context.Background(), id, "node1", "criticals"); err == nil {
		t.Fatal("command failure hidden")
	}
}
