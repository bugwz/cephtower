package clusterinspect

import (
	"context"
	"math/big"
	"sort"
	"strings"
	"time"
	"unicode"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

// RGWRealmUserCounts samples one registered zone per realm. It is not an
// inventory of every configured realm or an assertion of zone synchronization.
func (s *Service) RGWRealmUserCounts(ctx context.Context, clusterID uint64) (map[string]any, error) {
	return s.rgwRealmAggregate(ctx, clusterID, false)
}

// RGWRealmBucketUsage uses the same representative selection as user counts.
// Bucket, object and byte totals are exact decimal strings, not physical usage.
func (s *Service) RGWRealmBucketUsage(ctx context.Context, clusterID uint64) (map[string]any, error) {
	return s.rgwRealmAggregate(ctx, clusterID, true)
}

func (s *Service) rgwRealmAggregate(ctx context.Context, clusterID uint64, bucketUsage bool) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	started := time.Now().UTC()
	daemons, err := s.RGWDaemons(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	bad := func() error {
		return &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "registered RGW realm identities are incomplete or inconsistent"}
	}
	selected := map[string]RGWDaemon{}
	zoneRealms := map[string]string{}
	for _, daemon := range daemons {
		for _, id := range []string{daemon.RealmID, daemon.ZoneID} {
			if id == "" || len(id) > 256 || strings.ContainsFunc(id, unicode.IsControl) {
				return nil, bad()
			}
		}
		if realm, exists := zoneRealms[daemon.ZoneID]; exists && realm != daemon.RealmID {
			return nil, bad()
		}
		zoneRealms[daemon.ZoneID] = daemon.RealmID
		previous, exists := selected[daemon.RealmID]
		if !exists || daemon.ZoneID < previous.ZoneID || (daemon.ZoneID == previous.ZoneID && daemon.ServiceMapID < previous.ServiceMapID) {
			selected[daemon.RealmID] = daemon
		}
	}
	realms := make([]string, 0, len(selected))
	for realm := range selected {
		realms = append(realms, realm)
	}
	sort.Strings(realms)
	rows := make([]map[string]any, 0, len(realms))
	fields := []string{"user_count"}
	if bucketUsage {
		fields = []string{"bucket_count", "object_count", "size_actual_bytes"}
	}
	totals := map[string]*big.Int{}
	for _, field := range fields {
		totals[field] = new(big.Int)
	}
	for _, realm := range realms {
		daemon := selected[realm]
		var row map[string]any
		if bucketUsage {
			row, err = s.RGWBucketUsageInZone(ctx, clusterID, realm, daemon.ZoneID)
		} else {
			row, err = s.RGWUserCountInZone(ctx, clusterID, realm, daemon.ZoneID)
		}
		if err != nil {
			return nil, err
		}
		for _, field := range fields {
			value := new(big.Int)
			switch count := row[field].(type) {
			case int:
				if count < 0 {
					return nil, bad()
				}
				value.SetUint64(uint64(count))
			case string:
				if _, ok := value.SetString(count, 10); !ok || value.Sign() < 0 {
					return nil, bad()
				}
			default:
				return nil, bad()
			}
			totals[field].Add(totals[field], value)
			row[field] = value.String()
		}
		row["service_map_id"] = daemon.ServiceMapID
		rows = append(rows, row)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := map[string]any{"scope": "registered_realms", "source": "service_map+radosgw-admin", "selection": "lowest_zone_id_then_service_map_id", "realm_count": len(realms), "items": rows, "started_at": started, "observed_at": time.Now().UTC()}
	for _, field := range fields {
		result[field] = totals[field].String()
	}
	if bucketUsage {
		result["usage_category"] = "rgw.main"
	}
	return result, nil
}
