package clusterinspect

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWRealmUserCounts(t *testing.T) {
	s, runner, id := testInspection(t)
	registration := `{"services":{"rgw":{"daemons":{"3":{"metadata":{"id":"a","realm_id":"r1","zone_id":"z2"}},"2":{"metadata":{"id":"b","realm_id":"r1","zone_id":"z1"}},"1":{"metadata":{"id":"c","realm_id":"r1","zone_id":"z1"}},"4":{"metadata":{"id":"d","realm_id":"r2","zone_id":"z3"}}}}}}`
	fail := ""
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID == "rgw.daemons.read" {
			return executor.CommandResult{Stdout: []byte(registration)}, nil
		}
		args := strings.Join(spec.Args, " ")
		realm, zone := "r1", "z1"
		if strings.Contains(args, "--realm-id=r2") {
			realm, zone = "r2", "z3"
		}
		if !strings.Contains(args, "--zone-id="+zone) {
			t.Fatal("wrong representative", args)
		}
		raw := `["same-user"]`
		if spec.ID == "rgw.counts.user.zone" {
			raw = `{"id":"` + zone + `","realm_id":"` + realm + `"}`
		}
		exit := 0
		if realm == fail {
			exit = 2
		}
		return executor.CommandResult{Stdout: []byte(raw), ExitCode: exit}, nil
	}
	result, err := s.RGWRealmUserCounts(context.Background(), id)
	if err != nil || result["realm_count"] != 2 || result["user_count"] != "2" || len(runner.specs) != 5 {
		t.Fatal(result, err)
	}
	rows := result["items"].([]map[string]any)
	if rows[0]["service_map_id"] != "1" || rows[0]["zone_id"] != "z1" || rows[1]["realm_id"] != "r2" {
		t.Fatal("unstable selection", rows)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "same-user") {
		t.Fatal("identity leaked")
	}
	fail = "r2"
	if result, err = s.RGWRealmUserCounts(context.Background(), id); err == nil || result != nil {
		t.Fatal("partial aggregate", result)
	}
	fail = ""
	for _, raw := range []string{
		`{"services":{"rgw":{"daemons":{"1":{"metadata":{"id":"a","zone_id":"z"}}}}}}`,
		`{"services":{"rgw":{"daemons":{"1":{"metadata":{"id":"a","zone_id":"z","realm_id":"r1"}},"2":{"metadata":{"id":"b","zone_id":"z","realm_id":"r2"}}}}}}`,
	} {
		registration = raw
		before := len(runner.specs)
		if result, err = s.RGWRealmUserCounts(context.Background(), id); err == nil || result != nil || len(runner.specs) != before+1 {
			t.Fatal("invalid registrations read users")
		}
	}
	registration = `{"services":{}}`
	if result, err = s.RGWRealmUserCounts(context.Background(), id); err != nil || result["user_count"] != "0" || result["realm_count"] != 0 || len(result["items"].([]map[string]any)) != 0 {
		t.Fatal("empty registration scope", result, err)
	}
}
