package ceph

import (
	"context"
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

type rbdMirrorScheduleStatusWire struct {
	ScheduledImages []struct {
		ScheduleTime string `json:"schedule_time"`
		Image        string `json:"image"`
	} `json:"scheduled_images"`
}

func (p *NativeProvider) attachRBDMirrorSnapshotSchedules(ctx context.Context, access ClusterAccess, rows []Observation) {
	hasImages := false
	for _, row := range rows {
		if row.Kind == "rbd_image" {
			hasImages = true
			break
		}
	}
	if !hasImages {
		return
	}
	var schedules []rbdMirrorScheduleWire
	if !p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirror_schedule", []string{
		"mirror", "snapshot", "schedule", "list", "--recursive", "--format", "json",
	}, &schedules) || schedules == nil {
		return
	}

	nextRuns := map[string]string{}
	var status rbdMirrorScheduleStatusWire
	if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirror_schedule_status", []string{
		"mirror", "snapshot", "schedule", "status", "--format", "json",
	}, &status) {
		for _, item := range status.ScheduledImages {
			if name := strings.TrimSpace(item.Image); name != "" && strings.TrimSpace(item.ScheduleTime) != "" {
				nextRuns[name] = strings.TrimSpace(item.ScheduleTime)
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
	pool := strings.TrimSpace(candidate.Pool)
	namespace := strings.TrimSpace(candidate.Namespace)
	name := strings.TrimSpace(candidate.Image)
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
		if strings.ContainsAny(item.Interval, "\x00\r\n") || !rbdMirrorCollectedIntervalPattern.MatchString(strings.TrimSpace(item.Interval)) || len(item.StartTime) > 64 || strings.ContainsAny(item.StartTime, "\x00\r\n") {
			return false
		}
	}
	return true
}
