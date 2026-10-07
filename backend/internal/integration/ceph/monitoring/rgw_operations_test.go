package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRGWOperationCounters(t *testing.T) {
	for id, counter := range map[string]string{
		"rgw_get_bytes_total": "get_obj_bytes", "rgw_put_bytes_total": "put_obj_bytes",
		"rgw_copy_bytes_total": "copy_obj_bytes", "rgw_delete_bytes_total": "del_obj_bytes",
		"rgw_get_ops_total": "get_obj_ops", "rgw_put_ops_total": "put_obj_ops",
		"rgw_delete_ops_total": "del_obj_ops", "rgw_copy_ops_total": "copy_obj_ops",
		"rgw_list_objects_total": "list_obj_ops", "rgw_list_buckets_total": "list_buckets_ops",
		"rgw_delete_buckets_total": "del_bucket_ops",
	} {
		for _, history := range []bool{false, true} {
			client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"9007199254740993"]`
				if history {
					path, kind, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"9007199254740993"],[3660,"0"]]`
				}
				if r.URL.Path != path || r.URL.Query().Get("query") != "ceph_rgw_op_"+counter {
					t.Fatalf("wrong query: %s", r.URL)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[{"metric":{"instance_id":"rgw.a"},`+sample+`}]}}`), nil
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
			if len(result.Data.Result) != 1 || !strings.Contains(string(result.Data.Result[0]), `"instance_id":"rgw.a"`) || !strings.Contains(string(result.Data.Result[0]), `"9007199254740993"`) {
				t.Fatalf("lost raw counter: %+v", result)
			}
		}
	}
}

func TestRGWOperationLatencies(t *testing.T) {
	for id, operation := range map[string]string{
		"rgw_delete_latency_ms": "del_obj", "rgw_copy_latency_ms": "copy_obj",
		"rgw_list_objects_latency_ms": "list_obj", "rgw_list_buckets_latency_ms": "list_buckets",
		"rgw_delete_buckets_latency_ms": "del_bucket",
	} {
		for _, history := range []bool{false, true} {
			client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind, sample := "/prefix/api/v1/query", "vector", `"value":[3600,"NaN"]`
				if history {
					path, kind, sample = "/prefix/api/v1/query_range", "matrix", `"values":[[3600,"NaN"],[3660,"1.25"]]`
				}
				query := "1000 * sum(rate(ceph_rgw_op_" + operation + "_lat_sum[1m])) / sum(rate(ceph_rgw_op_" + operation + "_lat_count[1m]))"
				if r.URL.Path != path || r.URL.Query().Get("query") != query {
					t.Fatalf("wrong query: %s", r.URL)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[{"metric":{},`+sample+`}]}}`), nil
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
			if len(result.Data.Result) != 1 || !strings.Contains(string(result.Data.Result[0]), `"NaN"`) {
				t.Fatalf("unavailable latency changed: %+v", result)
			}
		}
	}
}
