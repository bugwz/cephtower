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
