package clusterinspect

import (
	"context"
	"encoding/json"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

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
