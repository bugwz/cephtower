package ceph

import (
	"context"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func (p *NativeProvider) collectPoolIO(ctx context.Context, access ClusterAccess) map[int64]*cephdomain.PoolClientIORate {
	var response []struct {
		ID   *int64                       `json:"pool_id"`
		Rate *cephdomain.PoolClientIORate `json:"client_io_rate"`
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.pool_io", []string{"osd", "pool", "stats", "--format", "json"}, &response) || response == nil {
		return nil
	}
	result := map[int64]*cephdomain.PoolClientIORate{}
	for _, row := range response {
		if row.ID == nil || *row.ID < 0 {
			return nil
		}
		if _, exists := result[*row.ID]; exists {
			return nil
		}
		result[*row.ID] = row.Rate
	}
	return result
}
