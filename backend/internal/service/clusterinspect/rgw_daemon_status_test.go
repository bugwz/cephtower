package clusterinspect

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRGWDaemonStatus(t *testing.T) {
	body := `{"rgw":{"12":{"status_stamp":"stamp","last_beacon":"beacon","status":{"current_sync":"idle","password":"private-password","json":"{\"secret_key\":\"private-key\",\"count\":9007199254740993}"}},"13":{"status_stamp":"other","last_beacon":"other","status":{}}}}`
	result, err := decodeRGWDaemonStatus([]byte(body), "12")
	if err != nil || result.ServiceMapID != "12" || result.StatusStamp != "stamp" || result.LastBeacon != "beacon" || result.Status["current_sync"] != "idle" {
		t.Fatal("wrong identity or fields", err)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "private-") || !strings.Contains(string(encoded), "9007199254740993") {
		t.Fatal("unsafe or lossy status")
	}
	for _, body := range []string{`null`, `[]`, `{"rgw":null}`, `{"rgw":{"12":null}}`, `{"rgw":{"12":{"status_stamp":"x","last_beacon":"y","status":{"json":"invalid"}}}}`, `{}`} {
		if _, err := decodeRGWDaemonStatus([]byte(body), "12"); err == nil {
			t.Fatal("invalid status accepted")
		}
	}
	if _, err := decodeRGWDaemonStatus([]byte(body), "missing"); err == nil {
		t.Fatal("missing identity accepted")
	}
	s, runner, id := testInspection(t)
	output := []byte(body)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.Mutating || spec.Binary != executor.BinaryCeph || !reflect.DeepEqual(spec.Args, []string{"service", "status", "--format", "json"}) {
			t.Fatal("wrong command")
		}
		return executor.CommandResult{Stdout: output}, nil
	}
	if _, err := s.RGWDaemonStatus(context.Background(), id, "12"); err != nil {
		t.Fatal(err)
	}
	for _, b := range output {
		if b != 0 {
			t.Fatal("output retained")
		}
	}
	before := len(runner.specs)
	if _, err := s.RGWDaemonStatus(context.Background(), id, "\n"); err == nil || len(runner.specs) != before {
		t.Fatal("invalid identity executed")
	}
	runner.run = func(executor.CommandSpec) (executor.CommandResult, error) {
		return executor.CommandResult{ExitCode: 1, Stdout: []byte("private-error")}, nil
	}
	if _, err := s.RGWDaemonStatus(context.Background(), id, "12"); err == nil || strings.Contains(err.Error(), "private-") {
		t.Fatal("unsafe failure")
	}
}
