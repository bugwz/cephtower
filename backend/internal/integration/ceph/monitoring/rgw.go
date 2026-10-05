package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// QueryRGWPerf reads exported snapshots, not Dashboard's manager-history rates.
// The mgr exporter labels RGW counters with its service-map id as instance_id.
func (c *Client) QueryRGWPerf(ctx context.Context, serviceMapID string) (PrometheusResult, error) {
	if serviceMapID == "" || len(serviceMapID) > 256 || strings.ContainsFunc(serviceMapID, unicode.IsControl) {
		return PrometheusResult{}, fmt.Errorf("invalid RGW service map identity")
	}
	quoted, _ := json.Marshal(serviceMapID)
	query := `{__name__=~"ceph_rgw_.*",instance_id=` + string(quoted) + `}`
	var result PrometheusResult
	err := c.get(ctx, "/api/v1/query?"+url.Values{"query": []string{query}}.Encode(), &result)
	if err != nil {
		return PrometheusResult{}, err
	}
	if err = validateMetricResult(result, "vector"); err != nil {
		return PrometheusResult{}, err
	}
	for _, raw := range result.Data.Result {
		var series struct {
			Metric map[string]string `json:"metric"`
			Value  []json.RawMessage `json:"value"`
		}
		if json.Unmarshal(raw, &series) != nil || series.Metric["instance_id"] != serviceMapID || !strings.HasPrefix(series.Metric["__name__"], "ceph_rgw_") || len(series.Value) != 2 {
			return PrometheusResult{}, fmt.Errorf("invalid RGW metric identity or sample")
		}
		var timestamp float64
		var value string
		if json.Unmarshal(series.Value[0], &timestamp) != nil || string(series.Value[0]) == "null" || json.Unmarshal(series.Value[1], &value) != nil || string(series.Value[1]) == "null" {
			return PrometheusResult{}, fmt.Errorf("invalid RGW metric sample")
		}
	}
	return result, nil
}
