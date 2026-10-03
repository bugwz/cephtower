package clusterinspect

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestHardwareFirmwareInventory(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"node1":{"bios":{"name":"System BIOS","version":"01.02","release_date":"2026-01-02","status":{"health":"OK","state":"Enabled"},"counter":18446744073709551615,"password":"fixture-secret"},"bmc":{}}}`
	result, err := s.Hardware(context.Background(), id, "node1", "firmwares")
	if err != nil {
		t.Fatal(err)
	}
	rows := result["items"].([]map[string]any)
	if len(rows) != 2 || rows[0]["version"] != "01.02" || rows[0]["name"] != "System BIOS" || rows[0]["release_date"] != "2026-01-02" || rows[0]["health"] != "OK" || rows[1]["health"] != nil || rows[1]["version"] != nil {
		t.Fatalf("%#v", result)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "fixture-secret") || !strings.Contains(rows[0]["details"].(string), "18446744073709551615") {
		t.Fatal("firmware details lost precision or leaked secrets")
	}
	if runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "hardware", "status", "--hostname", "node1", "--category", "firmwares", "--format", "json"}) {
		t.Fatal(runner.specs)
	}
	before := len(runner.specs)
	if _, err := s.Hardware(context.Background(), id, "", "firmwares"); err == nil || len(runner.specs) != before {
		t.Fatal("hostless firmware query executed")
	}
	for _, output := range []string{`null`, `[]`, `{"other":{}}`, `{"node1":null}`, `{"node1":{"bios":null}}`, `{"node1":{"bios":{"version":1}}}`, `{"node1":{"bios":{"status":{"health":false}}}}`, `{} {}`} {
		runner.output = output
		if result, err := s.Hardware(context.Background(), id, "node1", "firmwares"); err == nil || result != nil {
			t.Fatalf("accepted invalid response %s", output)
		}
	}
	for _, output := range []string{`{}`, `{"node1":{}}`} {
		runner.output = output
		result, err := s.Hardware(context.Background(), id, "node1", "firmwares")
		if err != nil || len(result["items"].([]map[string]any)) != 0 {
			t.Fatalf("empty inventory: %#v %v", result, err)
		}
	}
	runner.fail = true
	if _, err := s.Hardware(context.Background(), id, "node1", "firmwares"); err == nil {
		t.Fatal("command failure hidden")
	}
}
