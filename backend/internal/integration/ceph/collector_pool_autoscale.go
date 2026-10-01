package ceph

import (
	"context"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func (p *NativeProvider) collectPoolAutoscale(ctx context.Context, access ClusterAccess) map[int64]*cephdomain.PoolAutoscaleStatus {
	var response []struct {
		ID *int64 `json:"pool_id"`
		cephdomain.PoolAutoscaleStatus
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.pool_autoscale", []string{"osd", "pool", "autoscale-status", "--format", "json"}, &response) || response == nil {
		return nil
	}
	result := map[int64]*cephdomain.PoolAutoscaleStatus{}
	for _, row := range response {
		if row.ID == nil || *row.ID < 0 {
			return nil
		}
		if _, exists := result[*row.ID]; exists {
			return nil
		}
		for _, value := range []*float64{row.TargetRatio, row.EffectiveTargetRatio, row.Bias, row.LogicalUsed, row.RawUsedRate, row.ActualCapacityRatio, row.CapacityRatio} {
			if value != nil && *value < 0 {
				return nil
			}
		}
		status := row.PoolAutoscaleStatus
		result[*row.ID] = &status
	}
	return result
}
