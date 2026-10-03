package clusterinspect

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestHardwareNativeCategories(t *testing.T) {
	s, runner, id := testInspection(t)
	for _, category := range []string{"memory", "storage", "processors", "network", "power", "fans"} {
		runner.output = `{"node1":{"system1":{"part1":{"status":{"health":"OK","state":"Enabled"},"capacity_bytes":18446744073709551615},"part2":{}}}}`
		result, err := s.Hardware(context.Background(), id, "node1", category)
		if err != nil {
			t.Fatal(err)
		}
		rows := result["items"].([]map[string]any)
		if len(rows) != 2 || rows[0]["health"] != "OK" || rows[1]["health"] != nil || !strings.Contains(rows[0]["details"].(string), "18446744073709551615") {
			t.Fatalf("%+v", result)
		}
		spec := runner.specs[len(runner.specs)-1]
		if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"orch", "hardware", "status", "--hostname", "node1", "--category", category, "--format", "json"}) {
			t.Fatalf("%+v", spec)
		}
	}
	for _, output := range []string{`null`, `[]`, `{} {}`, `{"other":{}}`, `{"node1":null}`, `{"node1":{"s":null}}`, `{"node1":{"s":{"p":null}}}`, `{"node1":{"s":{"p":{"status":{"health":false}}}}}`} {
		runner.output = output
		if result, err := s.Hardware(context.Background(), id, "node1", "memory"); err == nil || result != nil {
			t.Fatalf("accepted %s", output)
		}
	}
	runner.output = `{}`
	result, err := s.Hardware(context.Background(), id, "node1", "memory")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	if !strings.Contains(string(raw), `"items":[]`) {
		t.Fatal(string(raw))
	}
	count := len(runner.specs)
	for _, args := range [][2]string{{"--help", "memory"}, {"node1", "shutdown"}, {"node1", ""}} {
		if _, err := s.Hardware(context.Background(), id, args[0], args[1]); err == nil {
			t.Fatal(args)
		}
	}
	if len(runner.specs) != count {
		t.Fatal("invalid request executed")
	}
	runner.fail = true
	if _, err := s.Hardware(context.Background(), id, "node1", "memory"); err == nil {
		t.Fatal("command error hidden")
	}
}

func TestHardwareDetailsRedactStructuredSecretsWithoutRounding(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"node1":{"sys":{"part":{"capacity_bytes":18446744073709551615,"status":{"health":"OK"},"vendor":{"password":"fixture-password","items":[{"access_key":"fixture-access"}],"client_key":{"nested":"fixture-nested"},"to\u006ben":"fixture-escaped"}}}}}`
	result, err := s.Hardware(context.Background(), id, "node1", "storage")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"fixture-password", "fixture-access", "fixture-nested", "fixture-escaped"} {
		if strings.Contains(string(raw), secret) {
			t.Fatal("hardware details leaked a structured credential")
		}
	}
	details := result["items"].([]map[string]any)[0]["details"].(string)
	if !json.Valid([]byte(details)) || !strings.Contains(details, "18446744073709551615") || !strings.Contains(details, "[REDACTED]") {
		t.Fatalf("invalid or rounded redacted details: %s", details)
	}
}

func TestHardwareClusterScopePreservesHostIdentity(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"node2":{"sys":{"part":{"status":{"health":"Warning"}}}},"node1":{"sys":{"part":{"status":{"health":"OK"}}}}}`
	result, err := s.Hardware(context.Background(), id, "", "memory")
	if err != nil {
		t.Fatal(err)
	}
	rows := result["items"].([]map[string]any)
	if len(rows) != 2 || rows[0]["host"] != "node1" || rows[1]["host"] != "node2" || rows[0]["id"] == rows[1]["id"] || result["host"] != "" {
		t.Fatalf("%+v", result)
	}
	if !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "hardware", "status", "--category", "memory", "--format", "json"}) || runner.specs[0].Mutating {
		t.Fatalf("%+v", runner.specs)
	}
	if _, err := s.Hardware(context.Background(), id, "node1", "memory"); err == nil {
		t.Fatal("scoped query accepted another host")
	}
	if _, err := s.Hardware(context.Background(), 0, "", "memory"); err == nil {
		t.Fatal("missing cluster accepted")
	}
}

func TestHardwareNativeAttributesAreProjectedPrecisely(t *testing.T) {
	s, runner, id := testInspection(t)
	for category, fields := range map[string]map[string]any{
		"memory":     {"description": "DIMM A1"},
		"storage":    {"description": "Disk 1", "model": "SSD", "capacity_bytes": "18446744073709551615", "protocol": "NVMe", "serial_number": "000123"},
		"processors": {"model": "CPU", "total_cores": "64", "total_threads": "128"},
		"network":    {"name": "NIC 1", "speed_mbps": "0"},
		"power":      {"name": "PSU 1", "model": "P1", "manufacturer": "Vendor"},
		"fans":       {"name": "Fan 1"},
	} {
		values, _ := json.Marshal(fields)
		// Exercise native numeric JSON, rather than an already-stringified fixture.
		raw := strings.ReplaceAll(string(values), `"18446744073709551615"`, `18446744073709551615`)
		raw = strings.ReplaceAll(raw, `"0"`, `0`)
		raw = strings.ReplaceAll(raw, `"64"`, `64`)
		raw = strings.ReplaceAll(raw, `"128"`, `128`)
		runner.output = `{"node1":{"sys":{"part":` + raw + `,"missing":{}}}}`
		result, err := s.Hardware(context.Background(), id, "node1", category)
		if err != nil {
			t.Fatal(err)
		}
		rows := result["items"].([]map[string]any)
		for field, want := range fields {
			if rows[1][field] != want || rows[0][field] != nil {
				t.Fatalf("%s %s: %#v", category, field, rows)
			}
		}
		if rows[1]["health"] != nil {
			t.Fatal("attributes fabricated health")
		}
	}
	for _, value := range []string{`true`, `{}`, `[]`} {
		runner.output = `{"node1":{"sys":{"part":{"capacity_bytes":` + value + `}}}}`
		if _, err := s.Hardware(context.Background(), id, "node1", "storage"); err == nil {
			t.Fatalf("accepted malformed attribute %s", value)
		}
	}
}
