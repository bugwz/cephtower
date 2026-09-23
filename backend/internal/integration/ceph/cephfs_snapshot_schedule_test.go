package ceph

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

type cephFSScheduleExecutor struct {
	calls   []executor.CommandSpec
	outputs map[string]string
	errors  map[string]error
}

func (e *cephFSScheduleExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	key := strings.Join(spec.Args, " ")
	if err := e.errors[key]; err != nil {
		return executor.CommandResult{}, err
	}
	output, exists := e.outputs[key]
	if !exists {
		return executor.CommandResult{}, errors.New("unexpected command: " + key)
	}
	return executor.CommandResult{Stdout: []byte(output)}, nil
}

func TestCollectCephFSSnapshotSchedulesDiscoversPathsThenReadsStatus(t *testing.T) {
	runner := &cephFSScheduleExecutor{outputs: map[string]string{
		"fs snap-schedule list / --recursive=true --fs=cephfs": "/projects 1h 24h\n/projects 1d 24h\n/archive 1w 4w\n",
		"fs snap-schedule status /projects --fs=cephfs --format=json": `[
			{"path":"/projects","schedule":"1h","start":"2026-09-23T00:00:00","active":true,"created_count":3,"retention":"24h"},
			{"path":"/projects","schedule":"1d","start":"2026-09-23T01:00:00","active":false,"created_count":1,"retention":"24h"}
		]`,
		"fs snap-schedule status /archive --fs=cephfs --format=json": `[{"path":"/archive","schedule":"1w","start":"2026-09-23T02:00:00","active":true,"pruned_count":2,"retention":"4w"}]`,
	}, errors: map[string]error{}}
	now := time.Unix(123, 0).UTC()
	provider := NativeProvider{Executor: runner}
	rows := provider.collectCephFSSnapshotSchedules(context.Background(), ClusterAccess{}, "cephfs", now)
	if len(rows) != 3 {
		t.Fatalf("rows = %#v", rows)
	}
	wantCommands := [][]string{
		{"fs", "snap-schedule", "list", "/", "--recursive=true", "--fs=cephfs"},
		{"fs", "snap-schedule", "status", "/projects", "--fs=cephfs", "--format=json"},
		{"fs", "snap-schedule", "status", "/archive", "--fs=cephfs", "--format=json"},
	}
	gotCommands := make([][]string, len(runner.calls))
	for index, call := range runner.calls {
		gotCommands[index] = call.Args
		if call.Binary != executor.BinaryCeph || call.Mutating {
			t.Fatalf("command %d = %+v", index, call)
		}
	}
	if !reflect.DeepEqual(gotCommands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", gotCommands, wantCommands)
	}
	first := rows[0]
	payload := first.Payload.(map[string]any)
	if first.ParentKey != "cephfs" || first.Status != "active" || payload["fs"] != "cephfs" || payload["path"] != "/projects" || textField(payload, "created_count") != "3" {
		t.Fatalf("first = %+v", first)
	}
	if rows[1].Status != "inactive" || rows[0].NaturalKey == rows[1].NaturalKey {
		t.Fatalf("schedule identities = %#v", rows[:2])
	}
}

func TestCollectCephFSSnapshotSchedulesTreatsJSONEmptyProbeAsAuthoritativeEmpty(t *testing.T) {
	plain := "fs snap-schedule list / --recursive=true --fs=cephfs"
	probe := plain + " --format=json"
	runner := &cephFSScheduleExecutor{
		outputs: map[string]string{probe: `{}`},
		errors:  map[string]error{plain: errors.New("ENOENT")},
	}
	provider := NativeProvider{Executor: runner}
	rows := provider.collectCephFSSnapshotSchedules(context.Background(), ClusterAccess{}, "cephfs", time.Time{})
	if len(rows) != 0 || len(runner.calls) != 2 || !reflect.DeepEqual(runner.calls[1].Args, []string{"fs", "snap-schedule", "list", "/", "--recursive=true", "--fs=cephfs", "--format=json"}) {
		t.Fatalf("rows=%#v calls=%+v", rows, runner.calls)
	}
}

func TestSchedulePathsMatchesDashboardPlainListDiscovery(t *testing.T) {
	want := []string{"/", "/projects", "/archive"}
	got, valid := schedulePaths("/ 1h 24h\n/projects 1h 24h\n/projects 1d 24h\n/archive 1w 4w")
	if !valid || !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
	if _, valid := schedulePaths("/projects 1h\ninvalid"); valid {
		t.Fatal("invalid list output accepted")
	}
}
