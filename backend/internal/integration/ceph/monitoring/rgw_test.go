package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
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

func TestRGWPerfHistory(t *testing.T) {
	end := time.Unix(7200, 0).UTC()
	for _, samples := range []string{`[[3600,"9007199254740993"],[7200,"NaN"]]`, `[]`, `[[3599,"1"]]`, `[[7201,"1"]]`, `[[3660,"1"],[3600,"2"]]`, `[[3600,"1"],[3600,"2"]]`, `[[3600,2]]`, `[[null,"1"]]`} {
		client, _ := New("https://prometheus.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			q := r.URL.Query()
			if r.URL.Path != "/api/v1/query_range" || q.Get("start") != "1970-01-01T01:00:00Z" || q.Get("end") != "1970-01-01T02:00:00Z" || q.Get("step") != "60" || q.Get("query") != `{__name__=~"ceph_rgw_.*",instance_id="1"}` {
				t.Fatal("invalid history scope")
			}
			return jsonResponse(200, `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"ceph_rgw_req","instance_id":"1"},"values":`+samples+`}]}}`), nil
		})})
		result, err := client.QueryRGWPerfHistory(context.Background(), "1", end)
		if samples == `[[3600,"9007199254740993"],[7200,"NaN"]]` {
			if err != nil || len(result.Data.Result) != 1 {
				t.Fatal("valid history rejected", err)
			}
		} else if err == nil {
			t.Fatal("invalid history accepted", samples)
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
