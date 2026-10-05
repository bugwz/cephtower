package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestRGWPerfScopedQuery(t *testing.T) {
	for _, id := range []string{"123", `id"},other="x`, `id\escaped`} {
		quoted, _ := json.Marshal(id)
		calls := 0
		payload, _ := json.Marshal(map[string]any{"status": "success", "data": map[string]any{"resultType": "vector", "result": []any{map[string]any{"metric": map[string]string{"__name__": "ceph_rgw_req", "instance_id": id, "job": "mgr"}, "value": []any{123, "9007199254740993"}}}}})
		client, _ := New("https://prometheus.example.test", "token", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Path != "/api/v1/query" || r.URL.Query().Get("query") != `{__name__=~"ceph_rgw_.*",instance_id=`+string(quoted)+`}` || r.Header.Get("Authorization") != "Bearer token" {
				t.Fatal("unscoped query")
			}
			return jsonResponse(200, string(payload)), nil
		})})
		result, err := client.QueryRGWPerf(context.Background(), id)
		if err != nil || len(result.Data.Result) != 1 || calls != 1 {
			t.Fatal("query failed", err)
		}
		var original PrometheusResult
		if json.Unmarshal(payload, &original) != nil || string(result.Data.Result[0]) != string(original.Data.Result[0]) {
			t.Fatal("sample precision or labels changed")
		}
		for _, bad := range []string{"", "bad\n"} {
			if _, err := client.QueryRGWPerf(context.Background(), bad); err == nil || calls != 1 {
				t.Fatal("invalid identity requested")
			}
		}
	}
}

func TestRGWPerfRejectsForeignSeries(t *testing.T) {
	for _, series := range []string{`{"metric":{"__name__":"ceph_rgw_req","instance_id":"other"},"value":[1,"2"]}`, `{"metric":{"__name__":"ceph_osd_op","instance_id":"1"},"value":[1,"2"]}`, `{"metric":{"__name__":"ceph_rgw_req","instance_id":"1"},"value":[1,2]}`, `{"metric":{"__name__":"ceph_rgw_req","instance_id":"1"},"value":[null,"2"]}`, `null`} {
		client, _ := New("https://prometheus.example.test", "", &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonResponse(200, `{"status":"success","data":{"resultType":"vector","result":[`+series+`]}}`), nil
		})})
		if _, err := client.QueryRGWPerf(context.Background(), "1"); err == nil {
			t.Fatal("invalid series accepted")
		}
	}
}
