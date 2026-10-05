package monitoring

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRGWOverviewQueries(t *testing.T) {
	queries := map[string]string{
		"rgw_request_rate":   "sum(rate(ceph_rgw_req[1m]))",
		"rgw_get_latency_ms": "(sum(rate(ceph_rgw_op_get_obj_lat_sum[1m])) / sum(rate(ceph_rgw_op_get_obj_lat_count[1m]))) * 1000",
		"rgw_put_latency_ms": "(sum(rate(ceph_rgw_op_put_obj_lat_sum[1m])) / sum(rate(ceph_rgw_op_put_obj_lat_count[1m]))) * 1000",
		"rgw_get_bytes_rate": "sum(rate(ceph_rgw_op_get_obj_bytes[1m]))",
		"rgw_put_bytes_rate": "sum(rate(ceph_rgw_op_put_obj_bytes[1m]))",
	}
	for id, query := range queries {
		for _, history := range []bool{false, true} {
			client, _ := New("https://prometheus.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind := "/api/v1/query", "vector"
				if history {
					path, kind = "/api/v1/query_range", "matrix"
				}
				if r.URL.Path != path || r.URL.Query().Get("query") != query {
					t.Fatal("wrong RGW query", id)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[]}}`), nil
			})})
			var err error
			if history {
				_, err = client.QueryRange(context.Background(), id, time.Unix(3600, 0), time.Unix(7200, 0), time.Minute)
			} else {
				_, err = client.Query(context.Background(), id, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
