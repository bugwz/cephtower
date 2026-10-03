package clusterinspect

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type RBDMirrorScheduledImage struct {
	Image        string `json:"image"`
	ScheduleTime string `json:"schedule_time"`
}

type RBDMirrorScheduleStatus struct {
	ScheduledImages []RBDMirrorScheduledImage `json:"scheduled_images"`
	ObservedAt      time.Time                 `json:"observed_at"`
}

func (s *Service) RBDMirrorScheduleStatus(ctx context.Context, clusterID uint64) (RBDMirrorScheduleStatus, error) {
	var result RBDMirrorScheduleStatus
	if err := s.readBinary(ctx, clusterID, "rbd.mirror.schedule.status", executor.BinaryRBD, []string{"mirror", "snapshot", "schedule", "status", "--format", "json"}, &result); err != nil {
		return RBDMirrorScheduleStatus{}, err
	}
	invalid := result.ScheduledImages == nil
	for _, item := range result.ScheduledImages {
		invalid = invalid || strings.TrimSpace(item.Image) == "" || strings.TrimSpace(item.ScheduleTime) == ""
	}
	if invalid {
		return RBDMirrorScheduleStatus{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RBD returned invalid mirror snapshot schedule status"}
	}
	result.ObservedAt = time.Now().UTC()
	return result, nil
}

type RBDMirrorSchedules struct {
	Schedules  json.RawMessage `json:"schedules"`
	ObservedAt time.Time       `json:"observed_at"`
}

func (s *Service) RBDMirrorSchedules(ctx context.Context, clusterID uint64) (RBDMirrorSchedules, error) {
	var raw json.RawMessage
	if err := s.readBinary(ctx, clusterID, "rbd.mirror.schedules", executor.BinaryRBD, []string{"mirror", "snapshot", "schedule", "list", "--recursive", "--format", "json"}, &raw); err != nil {
		return RBDMirrorSchedules{}, err
	}
	schedules, err := cephprovider.ParseRBDMirrorSchedules(raw)
	if err != nil {
		return RBDMirrorSchedules{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RBD returned invalid mirror snapshot schedules"}
	}
	return RBDMirrorSchedules{Schedules: schedules, ObservedAt: time.Now().UTC()}, nil
}
