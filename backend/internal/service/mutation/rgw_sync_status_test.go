package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

const syncReportFixture = "          realm realm-id (realm-a)\n      zonegroup group-id (group-a)\n           zone zone-id (zone-a)\n   current time 2026-10-05T00:00:00Z\nzonegroup features enabled: resharding\n  metadata sync no sync (zone is master)\n      data sync source: source-id (source-a)\n                       data is behind on 2 shards\n                       behind shards: [1,2]\n                       1 shards are recovering\n"

type syncReportExecutor struct {
	specs  []executor.CommandSpec
	result executor.CommandResult
	err    error
}

func (e *syncReportExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	return e.result, e.err
}

func TestReadRGWSyncStatus(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	runner := &syncReportExecutor{result: executor.CommandResult{Stdout: []byte(syncReportFixture), Stderr: []byte("private diagnostic")}}
	s.executor = runner
	got, err := s.ReadRGWSyncStatus(context.Background(), cluster, "zone-id", "zone-a")
	if err != nil || got != syncReportFixture {
		t.Fatalf("native report lost: %v", err)
	}
	if len(runner.specs) != 1 || runner.specs[0].Mutating || runner.specs[0].Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(runner.specs[0].Args, []string{"sync", "status", "--zone-id", "zone-id"}) || runner.specs[0].Timeout <= 0 || runner.specs[0].MaxOutput != 1<<20 {
		t.Fatal("unexpected command")
	}
	if strings.Trim(string(runner.result.Stdout), "\x00") != "" || strings.Trim(string(runner.result.Stderr), "\x00") != "" {
		t.Fatal("native buffers retained")
	}
	for _, args := range [][2]string{{"", "zone-a"}, {"zone-id", ""}, {"-id", "zone-a"}, {"zone-id", "bad\nname"}} {
		if _, err := s.ReadRGWSyncStatus(context.Background(), cluster, args[0], args[1]); err == nil {
			t.Fatal("invalid scope accepted")
		}
	}
	if _, err := s.ReadRGWSyncStatus(context.Background(), 0, "zone-id", "zone-a"); err == nil || len(runner.specs) != 1 {
		t.Fatal("invalid cluster executed command")
	}
	for _, commandErr := range []error{nil, errors.New("private diagnostic")} {
		runner.result = executor.CommandResult{Stdout: []byte(syncReportFixture), Stderr: []byte("private diagnostic"), ExitCode: 1}
		runner.err = commandErr
		got, err := s.ReadRGWSyncStatus(context.Background(), cluster, "zone-id", "zone-a")
		if err == nil || got != "" || strings.Contains(err.Error(), "private diagnostic") {
			t.Fatal("command failure accepted or exposed")
		}
	}
}

func TestValidateRGWSyncReport(t *testing.T) {
	for _, report := range []string{
		syncReportFixture,
		strings.Split(syncReportFixture, "      data sync")[0], // no source is not a replication verdict
		strings.ReplaceAll(syncReportFixture, "no sync (zone is master)", "syncing\n                full sync: 3/64 shards"),
		strings.ReplaceAll(syncReportFixture, "data is behind on 2 shards", "failed to fetch sync status: (13) Permission denied"),
	} {
		got, err := validateRGWSyncReport([]byte(report), "zone-id", "zone-a")
		if err != nil || got != report {
			t.Fatalf("report not preserved: %v", err)
		}
	}
	for _, report := range []string{"", "{}\n", strings.Repeat("x", 1<<20+1), syncReportFixture + "\xff", syncReportFixture + "\x00\n", strings.TrimSuffix(syncReportFixture, "\n"), strings.ReplaceAll(syncReportFixture, "zone-id", "wrong"), strings.ReplaceAll(syncReportFixture, "zone-a", "wrong"), strings.ReplaceAll(syncReportFixture, "metadata sync", "unknown"), strings.ReplaceAll(syncReportFixture, "zonegroup group-id", "unexpected group-id")} {
		got, err := validateRGWSyncReport([]byte(report), "zone-id", "zone-a")
		if err == nil || got != "" {
			t.Fatal("invalid report accepted")
		}
	}
}
