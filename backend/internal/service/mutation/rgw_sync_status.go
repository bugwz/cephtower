package mutation

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// ReadRGWSyncStatus returns the native text, not an inferred replication verdict.
// Even exit status zero may include remote-source errors or recovering shards.
func (s *Service) ReadRGWSyncStatus(ctx context.Context, clusterID uint64, zoneID, name string) (string, error) {
	if clusterID == 0 || !syncFlowToken(zoneID) || !syncFlowToken(name) {
		return "", invalid("cluster_id, zone_id and name are required")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return "", err
	}
	defer func() { access.ClientKey = "" }()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "rgw_zone.sync.status", Binary: executor.BinaryRGWAdmin,
		Args: []string{"sync", "status", "--zone-id", zoneID}, Timeout: time.Minute, MaxOutput: 1 << 20,
	})
	defer func() { clear(result.Stdout); clear(result.Stderr) }()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err != nil || result.ExitCode != 0 {
		return "", &cephdomain.ActionError{Code: "ceph_command_failed", Message: "Zone sync status read failed; check local zone configuration and cluster access"}
	}
	return validateRGWSyncReport(result.Stdout, zoneID, name)
}

func validateRGWSyncReport(raw []byte, zoneID, name string) (string, error) {
	failure := func() (string, error) {
		return "", &cephdomain.ActionError{Code: "capability_unavailable", Message: "Zone sync report unavailable, malformed or identity changed; refresh the zone inventory"}
	}
	if len(raw) == 0 || len(raw) > 1<<20 || !utf8.Valid(raw) || strings.ContainsRune(string(raw), '\x00') || raw[len(raw)-1] != '\n' {
		return failure()
	}
	lines := strings.Split(string(raw), "\n")
	// sync_status() in radosgw-admin prints these three identity headers first.
	// Keep all later feature, shard, lag and source-error text verbatim: their
	// presence is not proof of successful synchronization.
	if len(lines) < 5 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "realm ") ||
		!strings.HasPrefix(strings.TrimSpace(lines[1]), "zonegroup ") ||
		strings.TrimSpace(lines[2]) != "zone "+zoneID+" ("+name+")" {
		return failure()
	}
	for _, line := range lines[3:] {
		if strings.HasPrefix(strings.TrimSpace(line), "metadata sync ") {
			return string(raw), nil
		}
	}
	return failure()
}
