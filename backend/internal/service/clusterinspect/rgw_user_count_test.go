package clusterinspect

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserCount(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID != "rgw.counts.user" || spec.Binary != executor.BinaryRGWAdmin || spec.Mutating || !reflect.DeepEqual(spec.Args, []string{"user", "list", "--format", "json"}) {
			t.Fatal("incorrect unbounded native user enumeration", spec)
		}
		return executor.CommandResult{Stdout: []byte(`["alice","tenant$alice","account-user"]`)}, nil
	}
	result, err := s.RGWUserCount(context.Background(), id)
	if err != nil || result["user_count"] != 3 || result["source"] != "radosgw-admin" {
		t.Fatal(result, err)
	}
	if result["observed_at"].(time.Time).Before(result["started_at"].(time.Time)) {
		t.Fatal("invalid observation interval")
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "alice") {
		t.Fatal("user identities leaked into aggregate")
	}
	runner.run = nil
	runner.output = `[]`
	if result, err = s.RGWUserCount(context.Background(), id); err != nil || result["user_count"] != 0 {
		t.Fatal("empty list is not zero", result, err)
	}
	for _, raw := range []string{`null`, `{}`, `[null]`, `[1]`, `[""]`, `["a","a"]`, `["a\nb"]`, `{"keys":["a"],"truncated":true}`, `[] {}`} {
		runner.output = raw
		if result, err = s.RGWUserCount(context.Background(), id); err == nil || result != nil {
			t.Fatal("invalid list accepted", raw, result)
		}
	}
	for _, executionError := range []error{nil, errors.New("execution failed")} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			return executor.CommandResult{Stdout: []byte(`["a"]`), ExitCode: 2}, executionError
		}
		if result, err = s.RGWUserCount(context.Background(), id); err == nil || result != nil {
			t.Fatal("failed command produced a count")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := len(runner.specs)
	if result, err = s.RGWUserCount(ctx, id); !errors.Is(err, context.Canceled) || result != nil || len(runner.specs) != before {
		t.Fatal("cancelled read executed command", result, err)
	}
}
