package clusterinspect

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTelemetryStatusReadsNativeConfiguration(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"enabled":false,"channel_basic":true,"channel_ident":false,"interval":24,"last_opt_revision":3,"last_upload":null,"contact":null,"proxy":"http://user:secret@proxy:8080","url":"https://telemetry.ceph.com/report"}`
	result, err := s.TelemetryStatus(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	status := result["status"].(map[string]any)
	if status["enabled"] != false || status["interval"] != "24" || status["channel_ident"] != false || status["last_upload"] != nil || result["observed_at"] == nil {
		t.Fatalf("status=%v", result)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "user:secret") {
		t.Fatal("proxy credentials leaked")
	}
	if len(runner.specs) != 1 || runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"telemetry", "status", "--format", "json"}) {
		t.Fatalf("unexpected command: %+v", runner.specs)
	}
	for _, output := range []string{`{"enabled":true,"last_upload":0}`, `{"enabled":true,"last_upload":"Fri Oct 2 12:00:00 2026"}`, `{"enabled":false}`} {
		runner.output = output
		if _, err := s.TelemetryStatus(context.Background(), id); err != nil {
			t.Fatal(err)
		}
	}
	for _, output := range []string{`null`, `{}`, `[]`, `{"enabled":"false"}`, `{"enabled":true,"channel_basic":null}`, `{"enabled":true,"interval":1.5}`, `{"enabled":true,"last_opt_revision":-1}`, `{"enabled":true,"url":[]}`, `{"enabled":true,"last_upload":123}`} {
		runner.output = output
		if _, err := s.TelemetryStatus(context.Background(), id); err == nil {
			t.Fatalf("accepted malformed response %s", output)
		}
	}
	runner.fail = true
	if _, err := s.TelemetryStatus(context.Background(), id); err == nil {
		t.Fatal("command failure presented as disabled telemetry")
	}
}

func TestTelemetryReportCommandsAndPrecision(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `{"report":{"counter":18446744073709551615,"secret_key":"do-not-expose"},"device_report":{}}`
	for _, mode := range []string{"current", "preview"} {
		result, err := s.TelemetryReport(context.Background(), id, mode)
		if err != nil || result.Mode != mode || !strings.Contains(result.ReportJSON, "18446744073709551615") || strings.Contains(result.ReportJSON, "do-not-expose") || result.Message != "" {
			t.Fatalf("report=%+v err=%v", result, err)
		}
		command := "show-all"
		if mode == "preview" {
			command = "preview-all"
		}
		spec := runner.specs[len(runner.specs)-1]
		if spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"telemetry", command, "--format", "json"}) {
			t.Fatalf("unexpected command: %+v", spec)
		}
	}
	count := len(runner.specs)
	for _, mode := range []string{"", "send", "on", "off", "--help"} {
		if _, err := s.TelemetryReport(context.Background(), id, mode); err == nil {
			t.Fatal("invalid command accepted")
		}
	}
	if len(runner.specs) != count {
		t.Fatal("invalid mode executed a command")
	}
	runner.fail = true
	if _, err := s.TelemetryReport(context.Background(), id, "current"); err == nil {
		t.Fatal("failed command produced report")
	}
}

func TestTelemetryReportAvailabilityMessages(t *testing.T) {
	for mode, text := range map[string]string{
		"current": "Telemetry is off. Please consider opting-in with `ceph telemetry on`.\nPreview sample reports with `ceph telemetry preview`.",
		"preview": "Telemetry is up to date, see report with `ceph telemetry show`.",
	} {
		result, err := parseTelemetryReport([]byte(text), mode)
		if err != nil || result.Message != text || result.ReportJSON != "" {
			t.Fatalf("message=%+v err=%v", result, err)
		}
	}
	for _, data := range []string{"", "null", "[]", "failed", `{} {}`, `{"report":`} {
		if _, err := parseTelemetryReport([]byte(data), "preview"); err == nil {
			t.Fatalf("accepted invalid report: %s", data)
		}
	}
}
