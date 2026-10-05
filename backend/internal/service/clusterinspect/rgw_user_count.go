package clusterinspect

import (
	"context"
	"strings"
	"time"
	"unicode"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// RGWUserCount counts user metadata identities, not accounts or active sessions.
// Without max-entries or account selectors, user list exhausts native pagination.
func (s *Service) RGWUserCount(ctx context.Context, clusterID uint64) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now().UTC()
	var names []string
	if err := s.readBinary(ctx, clusterID, "rgw.counts.user", executor.BinaryRGWAdmin, []string{"user", "list", "--format", "json"}, &names); err != nil {
		return nil, err
	}
	bad := func() error {
		return &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "complete RGW user count is unavailable"}
	}
	if names == nil {
		return nil, bad()
	}
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] || strings.ContainsFunc(name, unicode.IsControl) {
			return nil, bad()
		}
		seen[name] = true
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]any{"user_count": len(names), "source": "radosgw-admin", "started_at": started, "observed_at": time.Now().UTC()}, nil
}
