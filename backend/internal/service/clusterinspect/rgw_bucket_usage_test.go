package clusterinspect

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
)

func TestRGWBucketUsageAggregate(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating {
			t.Fatal("unsafe command")
		}
		if spec.ID == "rgw.usage.list" {
			if !reflect.DeepEqual(spec.Args, []string{"metadata", "list", "bucket", "--format", "json"}) {
				t.Fatal("wrong enumeration")
			}
			return executor.CommandResult{Stdout: []byte(`["team/photos","photos"]`)}, nil
		}
		tenant := ""
		if len(spec.Args) == 6 {
			tenant = "team"
			if spec.Args[5] != "--tenant=team" {
				t.Fatal("tenant lost")
			}
		}
		if spec.Args[2] != "--bucket=photos" {
			t.Fatal("bucket lost")
		}
		return executor.CommandResult{Stdout: []byte(`{"bucket":"photos","tenant":"` + tenant + `","usage":{"rgw.main":{"num_objects":9007199254740993,"size_actual":18446744073709551615}}}`)}, nil
	}
	result, err := s.RGWBucketUsage(context.Background(), id)
	if err != nil || result["bucket_count"] != 2 || result["object_count"] != "18014398509481986" || result["size_actual_bytes"] != "36893488147419103230" {
		t.Fatal("inexact aggregate", result, err)
	}
}

func TestRGWBucketUsageRejectsIncompleteStatistics(t *testing.T) {
	for _, raw := range []string{`{}`, `{"bucket":"other","tenant":"","usage":{}}`, `{"bucket":"photos","tenant":"","usage":null}`, `{"bucket":"photos","tenant":"","usage":{"rgw.main":null}}`, `{"bucket":"photos","tenant":"","usage":{"rgw.main":{"num_objects":1,"size_actual":-1}}}`, `{"bucket":"photos","tenant":"","usage":{"rgw.main":{"num_objects":1,"size_actual":18446744073709551616}}}`} {
		s, runner, id := testInspection(t)
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if spec.ID == "rgw.usage.list" {
				return executor.CommandResult{Stdout: []byte(`["photos"]`)}, nil
			}
			return executor.CommandResult{Stdout: []byte(raw)}, nil
		}
		if result, err := s.RGWBucketUsage(context.Background(), id); err == nil || result != nil {
			t.Fatal("incomplete aggregate accepted", raw)
		}
	}
}

func TestRGWBucketUsageEmptyAndFailedReads(t *testing.T) {
	s, runner, id := testInspection(t)
	for _, list := range []string{`[]`, `["photos"]`} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if spec.ID == "rgw.usage.list" {
				return executor.CommandResult{Stdout: []byte(list)}, nil
			}
			return executor.CommandResult{Stdout: []byte(`{"bucket":"photos","tenant":"","usage":{}}`)}, nil
		}
		result, err := s.RGWBucketUsage(context.Background(), id)
		if err != nil || result["object_count"] != "0" || result["size_actual_bytes"] != "0" {
			t.Fatal("valid zero rejected", err)
		}
	}
	for _, list := range []string{`null`, `["photos","photos"]`, `["/photos"]`, `["a/b/c"]`, `[null]`} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if spec.ID == "rgw.usage.list" {
				return executor.CommandResult{Stdout: []byte(list)}, nil
			}
			return executor.CommandResult{Stdout: []byte(`{"bucket":"photos","tenant":"","usage":{}}`)}, nil
		}
		if result, err := s.RGWBucketUsage(context.Background(), id); err == nil || result != nil {
			t.Fatal("invalid names accepted")
		}
	}
	for _, stage := range []string{"rgw.usage.list", "rgw.usage.stats"} {
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			output := `["photos"]`
			if spec.ID == "rgw.usage.stats" {
				output = `{"bucket":"photos","tenant":"","usage":{}}`
			}
			exit := 0
			if spec.ID == stage {
				exit = 2
			}
			return executor.CommandResult{Stdout: []byte(output), ExitCode: exit}, nil
		}
		if result, err := s.RGWBucketUsage(context.Background(), id); err == nil || result != nil {
			t.Fatal("failed read returned total")
		}
	}
}
