package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
)

func (p *NativeProvider) collectOSDPerf(ctx context.Context, access ClusterAccess) map[int]*cephdomain.OSDPerfStats {
	var wire struct {
		Items []struct {
			ID    *int                     `json:"id"`
			Stats *cephdomain.OSDPerfStats `json:"perf_stats"`
		} `json:"osd_perf_infos"`
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.osd_perf", []string{"osd", "perf", "--format", "json"}, &wire) {
		return nil
	}
	invalid := func() map[int]*cephdomain.OSDPerfStats {
		markCollectionUnavailable(ctx, "collect.osd_perf")
		return nil
	}
	if wire.Items == nil {
		return invalid()
	}
	result := map[int]*cephdomain.OSDPerfStats{}
	for _, row := range wire.Items {
		if row.ID == nil || *row.ID < 0 || row.Stats == nil || result[*row.ID] != nil {
			return invalid()
		}
		for _, value := range []*float64{row.Stats.CommitLatencyMS, row.Stats.ApplyLatencyMS} {
			if value != nil && *value < 0 {
				return invalid()
			}
		}
		result[*row.ID] = row.Stats
	}
	return result
}
