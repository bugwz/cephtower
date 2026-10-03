package ceph

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type rbdMirrorScheduleExecutor struct {
	calls          []executor.CommandSpec
	scheduleOutput *string
	fail           bool
}

func (e *rbdMirrorScheduleExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	switch spec.ID {
	case "collect.rbd_mirror_schedule":
		if e.fail {
			return executor.CommandResult{}, fmt.Errorf("schedule query failed")
		}
		if e.scheduleOutput != nil {
			return executor.CommandResult{Stdout: []byte(*e.scheduleOutput)}, nil
		}
		return executor.CommandResult{Stdout: []byte(`[
          {"pool":"-","namespace":"-","image":"-","items":[{"interval":"1d","start_time":"00:15:00"}]},
          {"pool":"images","namespace":"-","image":"-","items":[{"interval":"12h","start_time":""}]},
          {"pool":"images","namespace":"team","image":"-","items":[{"interval":"2h","start_time":""}]},
          {"pool":"images","namespace":"team","image":"vm-1","items":[{"interval":"10m","start_time":""}]}
        ]`)}, nil
	case "collect.rbd_mirror_schedule_status":
		return executor.CommandResult{Stdout: []byte(`{"scheduled_images":[{"schedule_time":"2026-09-23 12:30:00","image":"images/team/vm-1"}]}`)}, nil
	default:
		return executor.CommandResult{}, fmt.Errorf("unexpected command %s", spec.ID)
	}
}

func TestAttachRBDMirrorSnapshotSchedulesUsesMostSpecificSchedule(t *testing.T) {
	runner := &rbdMirrorScheduleExecutor{}
	provider := NativeProvider{Executor: runner}
	rows := []Observation{{
		Kind: "rbd_image",
		Payload: cephdomain.RBDImage{
			ImagePath: "images/team/vm-1", Pool: "images", Namespace: "team", Name: "vm-1",
		},
	}}
	provider.attachRBDMirrorSnapshotSchedules(context.Background(), ClusterAccess{}, rows)
	image := rows[0].Payload.(cephdomain.RBDImage)
	if image.ScheduleInfo == nil || image.ScheduleInfo.Name != image.ImagePath || image.ScheduleInfo.InheritedFrom != "" || image.ScheduleInfo.NextRun != "2026-09-23 12:30:00" {
		t.Fatalf("schedule=%+v", image.ScheduleInfo)
	}
	if len(image.ScheduleInfo.Intervals) != 1 || image.ScheduleInfo.Intervals[0].Interval != "10m" {
		t.Fatalf("intervals=%+v", image.ScheduleInfo.Intervals)
	}
	want := [][]string{
		{"mirror", "snapshot", "schedule", "list", "--recursive", "--format", "json"},
		{"mirror", "snapshot", "schedule", "status", "--format", "json"},
	}
	for index, args := range want {
		if runner.calls[index].Binary != executor.BinaryRBD || !reflect.DeepEqual(runner.calls[index].Args, args) || runner.calls[index].Mutating {
			t.Fatalf("command %d=%+v", index, runner.calls[index])
		}
	}
}

func TestMirrorPoolSchedulesIncludeClusterAndExcludeOtherPools(t *testing.T) {
	runner := &rbdMirrorScheduleExecutor{}
	provider := NativeProvider{Executor: runner}
	rows := []Observation{
		{Kind: "rbd_mirroring", NaturalKey: "images", Payload: map[string]any{}},
		{Kind: "rbd_mirroring", NaturalKey: "archive", Payload: map[string]any{}},
	}
	provider.attachRBDMirrorSnapshotSchedules(context.Background(), ClusterAccess{}, rows)
	for index, count := range []int{4, 1} {
		payload := rows[index].Payload.(map[string]any)
		if payload["snapshot_schedules_status"] != "available" {
			t.Fatal(payload)
		}
		schedules := payload["snapshot_schedules"].([]rbdMirrorScheduleWire)
		if len(schedules) != count || schedules[0].Pool != "-" {
			t.Fatal(schedules)
		}
	}
	if len(runner.calls) != 2 {
		t.Fatalf("empty pools skipped schedule query: %v", runner.calls)
	}
}

func TestMirrorPoolSchedulesDistinguishEmptyAndUnavailable(t *testing.T) {
	for _, output := range []string{`[]`, `null`, `{}`, `[{"pool":"images","items":null}]`, `[{"pool":"images","items":[{"interval":"0m"}]}]`} {
		for _, fail := range []bool{false, true} {
			runner := &rbdMirrorScheduleExecutor{scheduleOutput: &output, fail: fail}
			provider := NativeProvider{Executor: runner}
			payload := map[string]any{}
			provider.attachRBDMirrorSnapshotSchedules(context.Background(), ClusterAccess{}, []Observation{{Kind: "rbd_mirroring", NaturalKey: "images", Payload: payload}})
			if output == `[]` && !fail {
				if payload["snapshot_schedules_status"] != "available" || len(payload["snapshot_schedules"].([]rbdMirrorScheduleWire)) != 0 {
					t.Fatal(payload)
				}
			} else if payload["snapshot_schedules_status"] != "unavailable" || payload["snapshot_schedules"] != nil {
				t.Fatal(payload)
			}
		}
	}
}

func TestRBDMirrorSnapshotScheduleInheritance(t *testing.T) {
	schedules := []rbdMirrorScheduleWire{
		{Pool: "-", Namespace: "-", Image: "-", Items: []cephdomain.RBDMirrorSnapshotScheduleItem{{Interval: "1d"}}},
		{Pool: "images", Namespace: "-", Image: "-", Items: []cephdomain.RBDMirrorSnapshotScheduleItem{{Interval: "12h"}}},
		{Pool: "images", Namespace: "team", Image: "-", Items: []cephdomain.RBDMirrorSnapshotScheduleItem{{Interval: "2h"}}},
		{Pool: "images", Namespace: "team", Image: "other", Items: []cephdomain.RBDMirrorSnapshotScheduleItem{{Interval: "5m"}}},
	}
	image := cephdomain.RBDImage{ImagePath: "images/team/vm-1", Pool: "images", Namespace: "team", Name: "vm-1"}
	schedule := rbdMirrorScheduleForImage(schedules, image)
	if schedule == nil || schedule.Name != "images/team/" || schedule.InheritedFrom != "namespace" || schedule.Intervals[0].Interval != "2h" {
		t.Fatalf("schedule=%+v", schedule)
	}
	image.Namespace = "other"
	image.ImagePath = "images/other/vm-1"
	schedule = rbdMirrorScheduleForImage(schedules, image)
	if schedule == nil || schedule.InheritedFrom != "pool" || schedule.Intervals[0].Interval != "12h" {
		t.Fatalf("pool schedule=%+v", schedule)
	}
	image.Pool = "archive"
	image.ImagePath = "archive/vm-1"
	schedule = rbdMirrorScheduleForImage(schedules, image)
	if schedule == nil || schedule.InheritedFrom != "cluster" || schedule.Intervals[0].Interval != "1d" {
		t.Fatalf("cluster schedule=%+v", schedule)
	}
}

func TestRBDMirrorSnapshotScheduleRejectsMalformedItems(t *testing.T) {
	image := cephdomain.RBDImage{ImagePath: "images/vm-1", Pool: "images", Name: "vm-1"}
	for _, items := range [][]cephdomain.RBDMirrorSnapshotScheduleItem{
		nil,
		{{Interval: ""}},
		{{Interval: "1h\n", StartTime: "00:00"}},
		{{Interval: "1h", StartTime: "00:00\n"}},
	} {
		if schedule := rbdMirrorScheduleForImage([]rbdMirrorScheduleWire{{Pool: "images", Namespace: "", Image: "vm-1", Items: items}}, image); schedule != nil {
			t.Fatalf("malformed schedule accepted: %+v", schedule)
		}
	}
}
