package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
)

type poolUsageWire struct {
	Stored      *uint64  `json:"stored"`
	BytesUsed   *uint64  `json:"bytes_used"`
	MaxAvail    *uint64  `json:"max_avail"`
	PercentUsed *float64 `json:"percent_used"`
}

func (p *NativeProvider) collectPoolUsage(ctx context.Context, access ClusterAccess) map[int64]poolUsageWire {
	var response struct {
		Pools []struct {
			ID    *int64        `json:"id"`
			Stats poolUsageWire `json:"stats"`
		} `json:"pools"`
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.pool_usage", []string{"df", "detail", "--format", "json"}, &response) || response.Pools == nil {
		return nil
	}
	result := map[int64]poolUsageWire{}
	for _, pool := range response.Pools {
		if pool.ID == nil || *pool.ID < 0 {
			return nil
		}
		if _, exists := result[*pool.ID]; exists {
			return nil
		}
		if pool.Stats.PercentUsed != nil {
			if *pool.Stats.PercentUsed < 0 || *pool.Stats.PercentUsed > 1 {
				return nil
			}
			percent := *pool.Stats.PercentUsed * 100
			pool.Stats.PercentUsed = &percent
		}
		result[*pool.ID] = pool.Stats
	}
	return result
}
