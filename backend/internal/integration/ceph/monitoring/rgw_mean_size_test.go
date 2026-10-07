package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRGWMeanPutSize(t *testing.T) {
	for _, history := range []bool{false, true} {
		for _, value := range []string{"0", "1024.5", "NaN", "+Inf"} {
			client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"`+value+`"]`
				if history {
					path, kind, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"`+value+`"]]`
				}
				if r.URL.Path != path || r.URL.Query().Get("query") != "sum(ceph_rgw_op_put_obj_bytes) / sum(ceph_rgw_op_put_obj_ops)" {
					t.Fatalf("wrong weighted query: %s", r.URL)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[{"metric":{},`+sample+`}]}}`), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			var result PrometheusResult
			if history {
				result, err = client.QueryRange(context.Background(), "rgw_put_mean_bytes", time.Unix(3600, 0), time.Unix(7200, 0), time.Minute)
			} else {
				result, err = client.Query(context.Background(), "rgw_put_mean_bytes", nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Data.Result) != 1 || !strings.Contains(string(result.Data.Result[0]), `"`+value+`"`) {
				t.Fatalf("changed sample: %+v", result)
			}
		}
	}
}
