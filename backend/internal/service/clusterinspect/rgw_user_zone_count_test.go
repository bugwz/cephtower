package clusterinspect

import (
	"context"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserCountInZone(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		kind, operation, raw := "user", "list", `["alice","tenant$alice"]`
		if spec.ID == "rgw.counts.user.zone" {
			kind, operation, raw = "zone", "get", `{"id":"zone-a","realm_id":"realm-a"}`
		}
		if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating || !reflect.DeepEqual(spec.Args, []string{kind, operation, "--format", "json", "--realm-id=realm-a", "--zone-id=zone-a"}) {
			t.Fatal("lost native scope", spec)
		}
		return executor.CommandResult{Stdout: []byte(raw)}, nil
	}
	result, err := s.RGWUserCountInZone(context.Background(), id, "realm-a", "zone-a")
	if err != nil || result["user_count"] != 2 || result["scope"] != "zone" || result["realm_id"] != "realm-a" || result["zone_id"] != "zone-a" || len(runner.specs) != 2 {
		t.Fatal("incorrect scoped count", result, err)
	}
	for _, raw := range []string{`{}`, `null`, `{"id":"other","realm_id":"realm-a"}`, `{"id":"zone-a","realm_id":"other"}`} {
		runner.run = nil
		runner.output = raw
		before := len(runner.specs)
		if result, err = s.RGWUserCountInZone(context.Background(), id, "realm-a", "zone-a"); err == nil || result != nil || len(runner.specs) != before+1 {
			t.Fatal("unverified zone enumerated users", raw)
		}
	}
	for _, ids := range [][2]string{{"", "z"}, {"r", ""}, {"r\n", "z"}, {"r", "z\x00"}} {
		before := len(runner.specs)
		if result, err = s.RGWUserCountInZone(context.Background(), id, ids[0], ids[1]); err == nil || result != nil || len(runner.specs) != before {
			t.Fatal("invalid identity executed")
		}
	}
	for _, failID := range []string{"rgw.counts.user.zone", "rgw.counts.user"} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			raw, exit := `[]`, 0
			if spec.ID == "rgw.counts.user.zone" {
				raw = `{"id":"zone-a","realm_id":"realm-a"}`
			}
			if spec.ID == failID {
				exit = 2
			}
			return executor.CommandResult{Stdout: []byte(raw), ExitCode: exit}, nil
		}
		if result, err = s.RGWUserCountInZone(context.Background(), id, "realm-a", "zone-a"); err == nil || result != nil {
			t.Fatal("failed scope returned count")
		}
	}
}
