package ceph

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

var rbdMirrorCollectedIntervalPattern = regexp.MustCompile(`^[1-9][0-9]*(?:m|h|d)$`)

type rbdMirrorScheduleWire struct {
	Pool      string                                     `json:"pool"`
	Namespace string                                     `json:"namespace"`
	Image     string                                     `json:"image"`
	Items     []cephdomain.RBDMirrorSnapshotScheduleItem `json:"items"`
}

func (s *rbdMirrorScheduleWire) UnmarshalJSON(data []byte) error {
	var wire struct {
		Pool      *string                                    `json:"pool"`
		Namespace *string                                    `json:"namespace"`
		Image     *string                                    `json:"image"`
		Items     []cephdomain.RBDMirrorSnapshotScheduleItem `json:"items"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Pool == nil || *wire.Pool == "" || wire.Namespace == nil || wire.Image == nil || *wire.Image == "" || wire.Items == nil {
		return fmt.Errorf("incomplete mirror schedule scope")
	}
	if (*wire.Pool == "-" && (*wire.Namespace != "-" || *wire.Image != "-")) || (*wire.Namespace == "-" && *wire.Image != "-") {
		return fmt.Errorf("invalid mirror schedule scope hierarchy")
	}
	*s = rbdMirrorScheduleWire{Pool: *wire.Pool, Namespace: *wire.Namespace, Image: *wire.Image, Items: wire.Items}
	return nil
}

type rbdMirrorScheduleStatusWire struct {
	ScheduledImages []struct {
		ScheduleTime string `json:"schedule_time"`
		Image        string `json:"image"`
	} `json:"scheduled_images"`
}

// ParseRBDMirrorSchedules validates native recursive-list output and returns
// only its typed schedule fields, for both live inspection and inventory use.
func ParseRBDMirrorSchedules(data []byte) (json.RawMessage, error) {
	var schedules []rbdMirrorScheduleWire
	if err := json.Unmarshal(data, &schedules); err != nil {
		return nil, err
	}
	if !validRBDMirrorSchedules(schedules) {
		return nil, fmt.Errorf("invalid mirror schedules")
	}
	return json.Marshal(schedules)
}

func validRBDMirrorSchedules(schedules []rbdMirrorScheduleWire) bool {
	if schedules == nil {
		return false
	}
	seen := map[[3]string]bool{}
	for _, schedule := range schedules {
		if schedule.Pool == "" || schedule.Items == nil || !validRBDMirrorScheduleItems(schedule.Items) {
			return false
		}
		scope := [3]string{schedule.Pool, schedule.Namespace, schedule.Image}
		if seen[scope] {
			return false
		}
		seen[scope] = true
	}
	return true
}

func (p *NativeProvider) attachRBDMirrorSnapshotSchedules(ctx context.Context, access ClusterAccess, rows []Observation) {
	hasResources := false
	for _, row := range rows {
		if row.Kind == "rbd_mirroring" || row.Kind == "rbd_namespace" {
			if payload, ok := row.Payload.(map[string]any); ok {
				payload["snapshot_schedules_status"] = "unavailable"
			}
		}
		if row.Kind == "rbd_image" || row.Kind == "rbd_mirroring" || row.Kind == "rbd_namespace" {
			hasResources = true
		}
	}
	if !hasResources {
		return
	}
	var schedules []rbdMirrorScheduleWire
	if !p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirror_schedule", []string{
		"mirror", "snapshot", "schedule", "list", "--recursive", "--format", "json",
	}, &schedules) || !validRBDMirrorSchedules(schedules) {
		return
	}
	for _, row := range rows {
		if row.Kind != "rbd_mirroring" && row.Kind != "rbd_namespace" {
			continue
		}
		payload, ok := row.Payload.(map[string]any)
		if !ok {
			continue
		}
		matching := make([]rbdMirrorScheduleWire, 0)
		pool, namespace := row.NaturalKey, ""
		if row.Kind == "rbd_namespace" {
			pool, namespace, _ = strings.Cut(row.NaturalKey, "/")
		}
		for _, schedule := range schedules {
			global := schedule.Pool == "-" && schedule.Namespace == "-" && schedule.Image == "-"
			inScope := schedule.Pool == pool
			if row.Kind == "rbd_namespace" {
				inScope = inScope && (schedule.Namespace == namespace || (schedule.Namespace == "-" && schedule.Image == "-"))
			}
			if inScope || global {
				matching = append(matching, schedule)
			}
		}
		payload["snapshot_schedules"] = matching
		payload["snapshot_schedules_status"] = "available"
	}

	nextRuns := map[string]string{}
	statusAvailable := false
	var status rbdMirrorScheduleStatusWire
	if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirror_schedule_status", []string{
		"mirror", "snapshot", "schedule", "status", "--format", "json",
	}, &status) {
		statusAvailable = status.ScheduledImages != nil
		for _, item := range status.ScheduledImages {
			if strings.TrimSpace(item.Image) == "" || strings.TrimSpace(item.ScheduleTime) == "" {
				statusAvailable = false
			}
		}
		for _, item := range status.ScheduledImages {
			if !statusAvailable {
				continue
			}
			// Native status is ordered by execution time; Dashboard also uses
			// the first exact image match rather than the last queued task.
			if _, exists := nextRuns[item.Image]; !exists {
				nextRuns[item.Image] = item.ScheduleTime
			}
		}
	}

	for index := range rows {
		if rows[index].Kind != "rbd_image" {
			continue
		}
		image, ok := rows[index].Payload.(cephdomain.RBDImage)
		if !ok {
			continue
		}
		if schedule := rbdMirrorScheduleForImage(schedules, image); schedule != nil {
			schedule.Status = "unavailable"
			if statusAvailable {
				schedule.Status = "available"
			}
			schedule.NextRun = nextRuns[image.ImagePath]
			image.ScheduleInfo = schedule
			rows[index].Payload = image
		}
	}
}

func rbdMirrorScheduleForImage(schedules []rbdMirrorScheduleWire, image cephdomain.RBDImage) *cephdomain.RBDMirrorSnapshotSchedule {
	bestScore := 0
	var best *cephdomain.RBDMirrorSnapshotSchedule
	for _, candidate := range schedules {
		score, name, inherited := rbdMirrorScheduleMatch(candidate, image)
		if score <= bestScore || len(candidate.Items) == 0 || !validRBDMirrorScheduleItems(candidate.Items) {
			continue
		}
		items := append([]cephdomain.RBDMirrorSnapshotScheduleItem(nil), candidate.Items...)
		bestScore = score
		best = &cephdomain.RBDMirrorSnapshotSchedule{Name: name, InheritedFrom: inherited, Intervals: items}
	}
	return best
}

func rbdMirrorScheduleMatch(candidate rbdMirrorScheduleWire, image cephdomain.RBDImage) (int, string, string) {
	pool := candidate.Pool
	namespace := candidate.Namespace
	name := candidate.Image
	if pool == "-" && namespace == "-" && name == "-" {
		return 1, "", "cluster"
	}
	if pool != image.Pool {
		return 0, "", ""
	}
	if namespace == "-" && name == "-" {
		return 2, pool + "/", "pool"
	}
	if namespace != image.Namespace {
		return 0, "", ""
	}
	if name == "-" {
		return 3, pool + "/" + namespace + "/", "namespace"
	}
	if name == image.Name {
		return 4, image.ImagePath, ""
	}
	return 0, "", ""
}

func validRBDMirrorScheduleItems(items []cephdomain.RBDMirrorSnapshotScheduleItem) bool {
	for _, item := range items {
		if !rbdMirrorCollectedIntervalPattern.MatchString(item.Interval) || len(item.StartTime) > 64 || strings.ContainsAny(item.StartTime, "\x00\r\n") {
			return false
		}
	}
	return true
}
