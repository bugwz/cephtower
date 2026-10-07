package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRGWIdentityLatencies(t *testing.T) {
	for _, scope := range []string{"bucket", "user"} {
		for id, operation := range map[string]string{"get": "get_obj", "put": "put_obj", "delete": "del_obj", "copy": "copy_obj", "list": "list_obj"} {
			for _, history := range []bool{false, true} {
				labels := `"` + scope + `":"same-name","tenant":"tenant-a","instance_id":"rgw.a"`
				client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					path, kind, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"NaN"]`
					if history {
						path, kind, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"NaN"],[3660,"1.25"]]`
					}
					prefix := "ceph_rgw_op_per_" + scope + "_" + operation
					if r.URL.Path != path || r.URL.Query().Get("query") != "1000 * rate("+prefix+"_lat_sum[1m]) / rate("+prefix+"_lat_count[1m])" {
						t.Fatalf("wrong query: %s", r.URL)
					}
					return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[{"metric":{`+labels+`},`+sample+`}]}}`), nil
				})})
				if err != nil {
					t.Fatal(err)
				}
				metric := "rgw_" + scope + "_" + id + "_latency_ms"
				var result PrometheusResult
				if history {
					result, err = client.QueryRange(context.Background(), metric, time.Unix(3600, 0), time.Unix(7200, 0), time.Minute)
				} else {
					result, err = client.Query(context.Background(), metric, nil)
				}
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Data.Result) != 1 || !strings.Contains(string(result.Data.Result[0]), labels) || !strings.Contains(string(result.Data.Result[0]), `"NaN"`) {
					t.Fatalf("lost identity or sample: %+v", result)
				}
			}
		}
	}
}
