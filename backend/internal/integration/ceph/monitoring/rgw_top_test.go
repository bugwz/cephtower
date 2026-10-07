package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRGWTopCounters(t *testing.T) {
	for _, scope := range []string{"bucket", "user"} {
		for _, op := range []string{"get", "put"} {
			for _, kind := range []string{"ops", "bytes"} {
				for _, history := range []bool{false, true} {
					labels := `"` + scope + `":"same-name","tenant":"a","instance_id":"rgw.1"`
					client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
						path, resultType, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"12"]`
						if history {
							path, resultType, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"12"]]`
						}
						if r.URL.Path != path || r.URL.Query().Get("query") != "topk(5, ceph_rgw_op_per_"+scope+"_"+op+"_obj_"+kind+")" {
							t.Fatalf("wrong top query: %s", r.URL)
						}
						return jsonResponse(200, `{"status":"success","data":{"resultType":"`+resultType+`","result":[{"metric":{`+labels+`},`+sample+`}]}}`), nil
					})})
					if err != nil {
						t.Fatal(err)
					}
					id := "rgw_" + scope + "_" + op + "_" + kind + "_top5"
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
						t.Fatalf("lost labels: %+v", result)
					}
				}
			}
		}
	}
}
