package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRGWSyncQueries(t *testing.T) {
	for id, query := range map[string]string{
		"rgw_sync_bytes_rate":      "sum by (source_zone) (rate(ceph_data_sync_from_zone_fetch_bytes_sum[1m]))",
		"rgw_sync_objects_rate":    "sum by (source_zone) (rate(ceph_data_sync_from_zone_fetch_bytes_count[1m]))",
		"rgw_sync_errors_rate":     "sum by (source_zone) (rate(ceph_data_sync_from_zone_fetch_errors[1m]))",
		"rgw_sync_delta_seconds":   "ceph_rgw_sync_delta_sync_delta",
		"rgw_sync_poll_latency_ms": "1000 * sum by (source_zone) (rate(ceph_data_sync_from_zone_poll_latency_sum[1m])) / sum by (source_zone) (rate(ceph_data_sync_from_zone_poll_latency_count[1m]))",
	} {
		for _, history := range []bool{false, true} {
			labels := `"source_zone":"west"`
			if id == "rgw_sync_delta_seconds" {
				labels = `"source_zone_id":"west-id","local_zone_id":"east-id","shard_id":"3","instance_id":"rgw.a"`
			}
			client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"1.25"]`
				if history {
					path, kind, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"1.25"]]`
				}
				if r.URL.Path != path || r.URL.Query().Get("query") != query {
					t.Fatalf("incorrect query for %s: %s", id, r.URL)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[{"metric":{`+labels+`},`+sample+`}]}}`), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			var result PrometheusResult
			if history {
				result, err = client.QueryRange(context.Background(), id, time.Unix(3600, 0), time.Unix(7200, 0), time.Minute)
			} else {
				result, err = client.Query(context.Background(), id, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Data.Result) != 1 || !strings.Contains(string(result.Data.Result[0]), labels) {
				t.Fatalf("missing source zone: %+v", result)
			}
		}
	}
}
