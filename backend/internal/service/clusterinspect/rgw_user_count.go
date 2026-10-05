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
	return s.rgwUserCount(ctx, clusterID, nil)
}

// RGWUserCountInZone verifies the exact configured zone before enumerating users.
// It does not discover realms or imply replication consistency between zones.
func (s *Service) RGWUserCountInZone(ctx context.Context, clusterID uint64, realmID, zoneID string) (map[string]any, error) {
	for _, id := range []string{realmID, zoneID} {
		if id == "" || len(id) > 256 || strings.ContainsFunc(id, unicode.IsControl) {
			return nil, invalid("realm_id and zone_id are required valid identities")
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now().UTC()
	selectors := []string{"--realm-id=" + realmID, "--zone-id=" + zoneID}
	var zone struct {
		ID      string `json:"id"`
		RealmID string `json:"realm_id"`
	}
	args := append([]string{"zone", "get", "--format", "json"}, selectors...)
	if err := s.readBinary(ctx, clusterID, "rgw.counts.user.zone", executor.BinaryRGWAdmin, args, &zone); err != nil {
		return nil, err
	}
	if zone.ID != zoneID || zone.RealmID != realmID {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RGW zone identity does not match the requested realm"}
	}
	result, err := s.rgwUserCount(ctx, clusterID, selectors)
	if err != nil {
		return nil, err
	}
	result["realm_id"], result["zone_id"] = realmID, zoneID
	result["scope"], result["started_at"] = "zone", started
	return result, nil
}

func (s *Service) rgwUserCount(ctx context.Context, clusterID uint64, selectors []string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now().UTC()
	var names []string
	args := append([]string{"user", "list", "--format", "json"}, selectors...)
	if err := s.readBinary(ctx, clusterID, "rgw.counts.user", executor.BinaryRGWAdmin, args, &names); err != nil {
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
