package ceph

import (
	"context"
	"strconv"
	"strings"

	"cephtower/backend/internal/integration/ceph/executor"
)

// PGMap::dump_pg_stats emits pgid and the full compound state for each PG.
// Reject an incomplete response as a whole rather than publishing partial counts.
func (p *NativeProvider) collectPoolPGStates(ctx context.Context, access ClusterAccess) map[int64]map[string]uint64 {
	var response struct {
		PGStats []struct {
			ID    string `json:"pgid"`
			State string `json:"state"`
		} `json:"pg_stats"`
	}
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.pool_pg_states", []string{"pg", "dump", "pgs_brief", "--format", "json"}, &response) || response.PGStats == nil {
		return nil
	}
	result := map[int64]map[string]uint64{}
	seen := map[string]bool{}
	for _, pg := range response.PGStats {
		pool, seed, ok := strings.Cut(pg.ID, ".")
		poolID, err := strconv.ParseInt(pool, 10, 64)
		_, seedErr := strconv.ParseUint(seed, 16, 32)
		if !ok || err != nil || poolID < 0 || seedErr != nil || strings.TrimSpace(pg.State) == "" || seen[pg.ID] {
			return nil
		}
		seen[pg.ID] = true
		if result[poolID] == nil {
			result[poolID] = map[string]uint64{}
		}
		result[poolID][pg.State]++
	}
	return result
}
