package clusterinspect

import (
	"context"
	"encoding/json"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// RGWTopologyCounts counts local configuration names, not live or remote sites.
// The native list commands exhaust their internal pagination before returning.
func (s *Service) RGWTopologyCounts(ctx context.Context, clusterID uint64) (map[string]any, error) {
	result := map[string]any{"source": "radosgw-admin", "started_at": time.Now().UTC()}
	for _, kind := range []string{"realm", "zonegroup", "zone"} {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var response map[string]json.RawMessage
		if err := s.readBinary(ctx, clusterID, "rgw.counts."+kind, executor.BinaryRGWAdmin, []string{kind, "list", "--format", "json"}, &response); err != nil {
			return nil, err
		}
		var names []string
		if json.Unmarshal(response[kind+"s"], &names) != nil || names == nil {
			return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RGW configuration list is invalid"}
		}
		seen := map[string]bool{}
		for _, name := range names {
			if name == "" || seen[name] {
				return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RGW configuration identity is invalid"}
			}
			seen[name] = true
		}
		result[kind+"_count"] = len(names)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result["observed_at"] = time.Now().UTC()
	return result, nil
}
