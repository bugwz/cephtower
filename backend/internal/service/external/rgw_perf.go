package external

import (
	"context"
	"strings"
	"time"
	"unicode"
)

// RGWPerf returns exported samples, not manager-history rates or daemon health.
func (s *Service) RGWPerf(ctx context.Context, clusterID uint64, serviceMapID string) (any, error) {
	if clusterID == 0 || serviceMapID == "" || len(serviceMapID) > 256 || strings.ContainsFunc(serviceMapID, unicode.IsControl) {
		return nil, failure("invalid_request", "cluster_id and a valid service_map_id are required", false)
	}
	api, err := s.monitoringClient(ctx, clusterID, "prometheus")
	if err != nil {
		return nil, err
	}
	result, err := api.QueryRGWPerf(ctx, serviceMapID)
	if err != nil {
		return nil, failure("prometheus_failed", "RGW performance snapshot could not be read", true)
	}
	return map[string]any{
		"service_map_id": serviceMapID, "source": "prometheus", "result_type": result.Data.ResultType,
		"series": result.Data.Result, "available": len(result.Data.Result) > 0, "observed_at": time.Now().UTC(),
	}, nil
}
