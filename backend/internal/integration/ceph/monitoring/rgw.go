package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// QueryRGWPerf reads exported snapshots, not Dashboard's manager-history rates.
// The mgr exporter labels RGW counters with its service-map id as instance_id.
func (c *Client) QueryRGWPerf(ctx context.Context, serviceMapID string) (PrometheusResult, error) {
	return c.queryRGWPerf(ctx, serviceMapID, nil)
}

// QueryRGWPerfHistory samples the last hour at one-minute evaluation intervals.
// Values remain raw exported metrics, not rates or original scrape timestamps.
func (c *Client) QueryRGWPerfHistory(ctx context.Context, serviceMapID string, end time.Time) (PrometheusResult, error) {
	if end.IsZero() {
		return PrometheusResult{}, fmt.Errorf("invalid history end time")
	}
	end = end.Truncate(time.Second)
	return c.queryRGWPerf(ctx, serviceMapID, &end)
}

func (c *Client) queryRGWPerf(ctx context.Context, serviceMapID string, end *time.Time) (PrometheusResult, error) {
	if serviceMapID == "" || len(serviceMapID) > 256 || strings.ContainsFunc(serviceMapID, unicode.IsControl) {
		return PrometheusResult{}, fmt.Errorf("invalid RGW service map identity")
	}
	quoted, _ := json.Marshal(serviceMapID)
	query := `{__name__=~"ceph_rgw_.*",instance_id=` + string(quoted) + `}`
	var result PrometheusResult
	path, resultType := "/api/v1/query", "vector"
	params := url.Values{"query": []string{query}}
	if end != nil {
		path, resultType = "/api/v1/query_range", "matrix"
		params.Set("start", end.Add(-time.Hour).UTC().Format(time.RFC3339Nano))
		params.Set("end", end.UTC().Format(time.RFC3339Nano))
		params.Set("step", "60")
	}
	err := c.get(ctx, path+"?"+params.Encode(), &result)
	if err != nil {
		return PrometheusResult{}, err
	}
	if err = validateMetricResult(result, resultType); err != nil {
		return PrometheusResult{}, err
	}
	for _, raw := range result.Data.Result {
		var series struct {
			Metric map[string]string   `json:"metric"`
			Value  []json.RawMessage   `json:"value"`
			Values [][]json.RawMessage `json:"values"`
		}
		if json.Unmarshal(raw, &series) != nil || series.Metric["instance_id"] != serviceMapID || !strings.HasPrefix(series.Metric["__name__"], "ceph_rgw_") {
			return PrometheusResult{}, fmt.Errorf("invalid RGW metric identity or sample")
		}
		samples := [][]json.RawMessage{series.Value}
		if end != nil {
			if len(series.Values) == 0 || len(series.Values) > 61 {
				return PrometheusResult{}, fmt.Errorf("invalid RGW history samples")
			}
			samples = series.Values
		}
		var previous float64
		for i, sample := range samples {
			var timestamp float64
			var value string
			if len(sample) != 2 || json.Unmarshal(sample[0], &timestamp) != nil || string(sample[0]) == "null" || json.Unmarshal(sample[1], &value) != nil || string(sample[1]) == "null" {
				return PrometheusResult{}, fmt.Errorf("invalid RGW metric sample")
			}
			if end != nil && (timestamp < float64(end.Add(-time.Hour).UnixMilli())/1000 || timestamp > float64(end.UnixMilli())/1000 || i > 0 && timestamp <= previous) {
				return PrometheusResult{}, fmt.Errorf("invalid RGW history timestamp")
			}
			previous = timestamp
		}
	}
	return result, nil
}
