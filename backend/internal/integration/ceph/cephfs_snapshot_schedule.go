package ceph

import (
	"context"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func (p *NativeProvider) collectCephFSSnapshotSchedules(ctx context.Context, access ClusterAccess, filesystem string, now time.Time) []Observation {
	if strings.TrimSpace(filesystem) == "" {
		markCollectionUnavailable(ctx, "collect.cephfs_schedule_list")
		return nil
	}
	listArgs := []string{"fs", "snap-schedule", "list", "/", "--recursive=true", "--fs=" + filesystem}
	result, err := p.runBinary(ctx, access, executor.BinaryCeph, "collect.cephfs_schedule_list", 45*time.Second, listArgs...)
	if err != nil {
		var emptyProbe map[string]any
		probeArgs := append(append([]string(nil), listArgs...), "--format=json")
		if probeErr := p.runInto(ctx, access, "collect.cephfs_schedule_list", probeArgs, &emptyProbe); probeErr == nil && len(emptyProbe) == 0 {
			return nil
		}
		markCollectionUnavailable(ctx, "collect.cephfs_schedule_list")
		return nil
	}
	paths, valid := schedulePaths(string(result))
	if !valid {
		markCollectionUnavailable(ctx, "collect.cephfs_schedule_list")
		return nil
	}
	if len(paths) == 0 {
		return nil
	}
	var rows []Observation
	for _, path := range paths {
		var schedules []map[string]any
		args := []string{"fs", "snap-schedule", "status", path, "--fs=" + filesystem, "--format=json"}
		if err := p.runInto(ctx, access, "collect.cephfs_schedule_status", args, &schedules); err != nil {
			continue
		}
		if schedules == nil {
			markCollectionUnavailable(ctx, "collect.cephfs_schedule_status")
			continue
		}
		for _, schedule := range schedules {
			schedulePath := textField(schedule, "path")
			repeat := textField(schedule, "schedule")
			start := textField(schedule, "start")
			active, activeOK := schedule["active"].(bool)
			if schedulePath == "" || repeat == "" || start == "" || !activeOK {
				markCollectionUnavailable(ctx, "collect.cephfs_schedule_status")
				continue
			}
			schedule["fs"] = filesystem
			schedule["filesystem"] = filesystem
			schedule["active"] = active
			key := opaquePair(filesystem, schedulePath+"\x00"+repeat+"\x00"+start)
			rows = append(rows, Observation{
				Kind:       "snapshot_schedule",
				NaturalKey: key,
				ParentKind: "filesystem",
				ParentKey:  filesystem,
				Name:       schedulePath + " " + repeat,
				Status:     map[bool]string{true: "active", false: "inactive"}[active],
				Source:     "ceph_cli",
				Payload:    schedule,
				ObservedAt: now,
			})
		}
	}
	return rows
}

func schedulePaths(output string) ([]string, bool) {
	seen := map[string]struct{}{}
	var paths []string
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "/") {
			return nil, false
		}
		if _, exists := seen[fields[0]]; exists {
			continue
		}
		seen[fields[0]] = struct{}{}
		paths = append(paths, fields[0])
	}
	return paths, true
}
