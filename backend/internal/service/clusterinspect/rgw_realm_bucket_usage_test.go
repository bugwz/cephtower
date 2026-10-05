package clusterinspect

import (
	"context"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWRealmBucketUsage(t *testing.T) {
	s, runner, id := testInspection(t)
	registration := `{"services":{"rgw":{"daemons":{"1":{"metadata":{"id":"a","realm_id":"r1","zone_id":"z1"}},"2":{"metadata":{"id":"b","realm_id":"r1","zone_id":"z1"}},"3":{"metadata":{"id":"c","realm_id":"r2","zone_id":"z2"}}}}}}`
	failStage, missingUsage := "", false
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID == "rgw.daemons.read" {
			return executor.CommandResult{Stdout: []byte(registration)}, nil
		}
		realm, zone := "r1", "z1"
		args := strings.Join(spec.Args, " ")
		if strings.Contains(args, "--realm-id=r2") {
			realm, zone = "r2", "z2"
		}
		if !strings.Contains(args, "--zone-id="+zone) {
			t.Fatal("lost representative zone", args)
		}
		raw := ""
		switch spec.ID {
		case "rgw.usage.zone":
			raw = `{"id":"` + zone + `","realm_id":"` + realm + `"}`
		case "rgw.usage.list":
			raw = `["photos"]`
		case "rgw.usage.stats":
			raw = `{"bucket":"photos","tenant":"","usage":{"rgw.main":{"num_objects":9007199254740993,"size_actual":18446744073709551615}}}`
			if missingUsage && realm == "r2" {
				raw = `{"bucket":"photos","tenant":""}`
			}
		default:
			t.Fatal("unexpected command", spec.ID)
		}
		exit := 0
		if realm == "r2" && failStage == spec.ID {
			exit = 2
		}
		return executor.CommandResult{Stdout: []byte(raw), ExitCode: exit}, nil
	}
	result, err := s.RGWRealmBucketUsage(context.Background(), id)
	if err != nil || result["bucket_count"] != "2" || result["object_count"] != "18014398509481986" || result["size_actual_bytes"] != "36893488147419103230" || result["realm_count"] != 2 || result["usage_category"] != "rgw.main" || len(runner.specs) != 7 {
		t.Fatal("inexact cross realm aggregate", result, err)
	}
	rows := result["items"].([]map[string]any)
	if rows[0]["service_map_id"] != "1" || rows[0]["bucket_count"] != "1" || rows[1]["realm_id"] != "r2" {
		t.Fatal("incorrect representatives", rows)
	}
	if _, exists := result["user_count"]; exists {
		t.Fatal("unrequested user aggregate")
	}
	for _, stage := range []string{"rgw.usage.zone", "rgw.usage.list", "rgw.usage.stats"} {
		failStage = stage
		if result, err = s.RGWRealmBucketUsage(context.Background(), id); err == nil || result != nil {
			t.Fatal("partial total on failed later realm", stage)
		}
	}
	failStage, missingUsage = "", true
	if result, err = s.RGWRealmBucketUsage(context.Background(), id); err == nil || result != nil {
		t.Fatal("missing usage produced total")
	}
	registration = `{"services":{}}`
	if result, err = s.RGWRealmBucketUsage(context.Background(), id); err != nil || result["bucket_count"] != "0" || result["object_count"] != "0" || result["size_actual_bytes"] != "0" {
		t.Fatal("empty registration scope", result, err)
	}
}
