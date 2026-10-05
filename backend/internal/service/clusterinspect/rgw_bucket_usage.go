package clusterinspect

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"time"
	"unicode"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// RGWBucketUsage sums rgw.main statistics over one complete enumeration.
// Reads are sequential, not an atomic cluster snapshot; remote/indexless
// buckets without usage make the aggregate unavailable rather than zero.
func (s *Service) RGWBucketUsage(ctx context.Context, clusterID uint64) (map[string]any, error) {
	return s.rgwBucketUsage(ctx, clusterID, nil)
}

// RGWBucketUsageInZone keeps enumeration and every bucket statistic in the
// same explicitly verified zone. It does not claim cross-zone consistency.
func (s *Service) RGWBucketUsageInZone(ctx context.Context, clusterID uint64, realmID, zoneID string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	started := time.Now().UTC()
	selectors, err := s.rgwVerifiedZoneSelectors(ctx, clusterID, "rgw.usage.zone", realmID, zoneID)
	if err != nil {
		return nil, err
	}
	result, err := s.rgwBucketUsage(ctx, clusterID, selectors)
	if err != nil {
		return nil, err
	}
	result["realm_id"], result["zone_id"] = realmID, zoneID
	result["scope"], result["started_at"] = "zone", started
	return result, nil
}

func (s *Service) rgwBucketUsage(ctx context.Context, clusterID uint64, selectors []string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	started := time.Now().UTC()
	bad := func() error {
		return &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "complete RGW bucket usage is unavailable"}
	}
	var names []string
	listArgs := append([]string{"metadata", "list", "bucket", "--format", "json"}, selectors...)
	if err := s.readBinary(ctx, clusterID, "rgw.usage.list", executor.BinaryRGWAdmin, listArgs, &names); err != nil {
		return nil, err
	}
	if names == nil {
		return nil, bad()
	}
	seen := map[string]bool{}
	objects, bytes := new(big.Int), new(big.Int)
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if name == "" || seen[name] || strings.ContainsFunc(name, unicode.IsControl) {
			return nil, bad()
		}
		seen[name] = true
		tenant, bucket := "", name
		if prefix, suffix, found := strings.Cut(name, "/"); found {
			tenant, bucket = prefix, suffix
			if tenant == "" {
				return nil, bad()
			}
		}
		if bucket == "" || strings.Contains(bucket, "/") {
			return nil, bad()
		}
		args := []string{"bucket", "stats", "--bucket=" + bucket, "--format", "json"}
		if tenant != "" {
			args = append(args, "--tenant="+tenant)
		}
		args = append(args, selectors...)
		var stats struct {
			Bucket *string                    `json:"bucket"`
			Tenant *string                    `json:"tenant"`
			Usage  map[string]json.RawMessage `json:"usage"`
		}
		if err := s.readBinary(ctx, clusterID, "rgw.usage.stats", executor.BinaryRGWAdmin, args, &stats); err != nil {
			return nil, err
		}
		if stats.Bucket == nil || *stats.Bucket != bucket || stats.Tenant == nil || *stats.Tenant != tenant || stats.Usage == nil {
			return nil, bad()
		}
		if raw, exists := stats.Usage["rgw.main"]; exists {
			var counters map[string]json.RawMessage
			if json.Unmarshal(raw, &counters) != nil || counters == nil {
				return nil, bad()
			}
			for _, field := range []struct {
				name  string
				total *big.Int
			}{{"num_objects", objects}, {"size_actual", bytes}} {
				text := string(counters[field.name])
				if text == "" || len(text) > 20 {
					return nil, bad()
				}
				for _, c := range text {
					if c < '0' || c > '9' {
						return nil, bad()
					}
				}
				value, ok := new(big.Int).SetString(text, 10)
				if !ok || value.BitLen() > 64 {
					return nil, bad()
				}
				field.total.Add(field.total, value)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]any{"bucket_count": len(names), "object_count": objects.String(), "size_actual_bytes": bytes.String(), "usage_category": "rgw.main", "source": "radosgw-admin", "started_at": started, "observed_at": time.Now().UTC()}, nil
}
