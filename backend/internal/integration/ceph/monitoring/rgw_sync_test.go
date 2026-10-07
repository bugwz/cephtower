package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRGWSyncQueries(t *testing.T) {
	for id, counter := range map[string]string{
		"rgw_sync_bytes_rate":   "ceph_data_sync_from_zone_fetch_bytes_sum",
		"rgw_sync_objects_rate": "ceph_data_sync_from_zone_fetch_bytes_count",
		"rgw_sync_errors_rate":  "ceph_data_sync_from_zone_fetch_errors",
	} {
		for _, history := range []bool{false, true} {
			client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"1.25"]`
				if history {
					path, kind, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"1.25"]]`
				}
				if r.URL.Path != path || r.URL.Query().Get("query") != "sum by (source_zone) (rate("+counter+"[1m]))" {
					t.Fatalf("incorrect query for %s: %s", id, r.URL)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[{"metric":{"source_zone":"west"},`+sample+`}]}}`), nil
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
			if len(result.Data.Result) != 1 || !strings.Contains(string(result.Data.Result[0]), `"source_zone":"west"`) {
				t.Fatalf("missing source zone: %+v", result)
			}
		}
	}
}
