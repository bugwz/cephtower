package external

import (
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestMonHistoryScope(t *testing.T) {
	for _, known := range []bool{true, false} {
		fsid := "cluster-a"
		var stored *string
		if known {
			stored = &fsid
		}
		s, endpoints, cluster := externalTestServiceWithFSID(t, stored)
		ctx := context.Background()
		if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "prometheus", URL: "https://prometheus.example.test"}); err != nil {
			t.Fatal(err)
		}
		calls := 0
		name := "a\"} or up{job=\"x"
		s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			want := "ceph_mon_num_sessions{cluster=" + strconv.Quote(fsid) + ",ceph_daemon=" + strconv.Quote("mon."+name) + "}"
			if r.URL.Path != "/api/v1/query_range" || r.URL.Query().Get("query") != want || r.URL.Query().Get("step") != "30" {
				t.Fatalf("unexpected request %s", r.URL)
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"status":"success","data":{"resultType":"matrix","result":[]}}`)), Request: r}, nil
		})
		query := url.Values{"metric_id": {"mon_sessions"}, "mon_name": {name}, "start": {"2026-01-01T00:00:00Z"}, "end": {"2026-01-01T01:00:00Z"}, "step": {"30s"}}
		result, err := s.readMetric(ctx, cluster.ID, "metric/range", query)
		if known {
			if err != nil || calls != 1 {
				t.Fatalf("result=%v err=%v calls=%d", result, err, calls)
			}
			meta := result.(map[string]any)["meta"].(map[string]any)
			if meta["cluster_fsid"] != fsid || meta["mon_name"] != name {
				t.Fatalf("meta=%v", meta)
			}
		} else if err == nil || result != nil || calls != 0 {
			t.Fatalf("unscoped history: %v %v %d", result, err, calls)
		}
		before := calls
		for _, key := range []string{"metric/query", "metric/range"} {
			invalid := url.Values{}
			for k, values := range query {
				invalid[k] = append([]string(nil), values...)
			}
			invalid.Del("mon_name")
			if _, err := s.readMetric(ctx, cluster.ID, key, invalid); err == nil {
				t.Fatal("missing MON accepted")
			}
		}
		if calls != before {
			t.Fatal("invalid request reached upstream")
		}
	}
}
