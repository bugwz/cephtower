package clusterinspect

import (
	"context"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWBucketUsageInZone(t *testing.T) {
	s, runner, id := testInspection(t)
	failID := ""
	zone := `{"id":"z","realm_id":"r"}`
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		var args []string
		raw := zone
		switch spec.ID {
		case "rgw.usage.zone":
			args = []string{"zone", "get", "--format", "json"}
		case "rgw.usage.list":
			args = []string{"metadata", "list", "bucket", "--format", "json"}
			raw = `["team/photos"]`
		case "rgw.usage.stats":
			args = []string{"bucket", "stats", "--bucket=photos", "--format", "json", "--tenant=team"}
			raw = `{"bucket":"photos","tenant":"team","usage":{"rgw.main":{"num_objects":9007199254740993,"size_actual":18446744073709551615}}}`
		default:
			t.Fatal("unexpected command", spec.ID)
		}
		args = append(args, "--realm-id=r", "--zone-id=z")
		if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating || !reflect.DeepEqual(spec.Args, args) {
			t.Fatal("lost zone or tenant selector", spec)
		}
		exit := 0
		if spec.ID == failID {
			exit = 2
		}
		return executor.CommandResult{Stdout: []byte(raw), ExitCode: exit}, nil
	}
	result, err := s.RGWBucketUsageInZone(context.Background(), id, "r", "z")
	if err != nil || result["object_count"] != "9007199254740993" || result["size_actual_bytes"] != "18446744073709551615" || result["realm_id"] != "r" || result["zone_id"] != "z" || result["scope"] != "zone" {
		t.Fatal(result, err)
	}
	for _, stage := range []string{"rgw.usage.zone", "rgw.usage.list", "rgw.usage.stats"} {
		failID = stage
		if result, err = s.RGWBucketUsageInZone(context.Background(), id, "r", "z"); err == nil || result != nil {
			t.Fatal("failed stage returned aggregate", stage)
		}
	}
	failID = ""
	zone = `{"id":"z","realm_id":"other"}`
	before := len(runner.specs)
	if result, err = s.RGWBucketUsageInZone(context.Background(), id, "r", "z"); err == nil || result != nil || len(runner.specs) != before+1 {
		t.Fatal("unverified zone enumerated buckets")
	}
	before = len(runner.specs)
	if result, err = s.RGWBucketUsageInZone(context.Background(), id, "", "z"); err == nil || result != nil || len(runner.specs) != before {
		t.Fatal("invalid zone scope executed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err = s.RGWBucketUsageInZone(ctx, id, "r", "z"); err == nil || result != nil || len(runner.specs) != before {
		t.Fatal("cancelled zone scope executed")
	}
}
