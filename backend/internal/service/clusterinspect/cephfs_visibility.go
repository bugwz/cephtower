package clusterinspect

import (
	"context"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
)

type SubvolumeVisibility struct {
	Filesystem string    `json:"filesystem"`
	Subvolume  string    `json:"subvolume"`
	Group      string    `json:"group"`
	Visible    bool      `json:"visible"`
	ObservedAt time.Time `json:"observed_at"`
}

func (s *Service) SubvolumeSnapshotVisibility(ctx context.Context, clusterID uint64, fs, subvolume, group string) (SubvolumeVisibility, error) {
	if clusterID == 0 || !cephFSNamePattern.MatchString(fs) || !cephFSNamePattern.MatchString(subvolume) || (group != "" && !cephFSNamePattern.MatchString(group)) {
		return SubvolumeVisibility{}, invalid("valid cluster, filesystem, subvolume and group are required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return SubvolumeVisibility{}, err
	}
	args := []string{"fs", "subvolume", "snapshot_visibility", "get", fs, subvolume}
	if group != "" && group != "_nogroup" {
		args = append(args, "--group_name", group)
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "subvolume.snapshot_visibility.get", Binary: executor.BinaryCeph, Args: args, Timeout: 20 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return SubvolumeVisibility{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: security.Redact(err.Error()), Retryable: true}
	}
	value := strings.TrimSpace(string(result.Stdout))
	if value != "0" && value != "1" {
		return SubvolumeVisibility{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned invalid snapshot visibility"}
	}
	if group == "" {
		group = "_nogroup"
	}
	return SubvolumeVisibility{Filesystem: fs, Subvolume: subvolume, Group: group, Visible: value == "1", ObservedAt: time.Now().UTC()}, nil
}
