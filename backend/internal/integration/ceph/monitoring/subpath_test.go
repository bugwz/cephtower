package monitoring

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMonitoringPreservesServiceRoot(t *testing.T) {
	for _, prefix := range []string{"", "/", "/monitor", "/monitor/", "/proxy/ceph%20cluster"} {
		t.Run(prefix, func(t *testing.T) {
			calls := 0
			paths := []string{"/api/search", "/api/v1/query", "/api/v1/rules", "/api/v2/alerts", "/api/v2/silences", "/api/v2/silences", "/api/v2/silence/silence%20id"}
			methods := []string{"GET", "GET", "GET", "GET", "GET", "POST", "DELETE"}
			bodies := []string{`[]`, `{"status":"success","data":{"resultType":"vector","result":[]}}`, `{"status":"success","data":{"groups":[]}}`, `[]`, `[]`, `{"silenceID":"new"}`, ``}
			client, err := New("https://monitor.example.test"+prefix, "secret", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				index := calls
				calls++
				if index >= len(paths) {
					t.Fatal("unexpected extra request")
				}
				if r.URL.EscapedPath() != strings.TrimSuffix(prefix, "/")+paths[index] || r.Method != methods[index] || r.URL.Host != "monitor.example.test" || r.Header.Get("Authorization") != "Bearer secret" {
					t.Fatalf("incorrect service request %s %s", r.Method, r.URL)
				}
				if index == 0 && (r.URL.Query().Get("type") != "dash-db" || r.URL.Query().Get("page") != "1") {
					t.Fatal("lost dashboard query")
				}
				if index == 1 && r.URL.Query().Get("query") != "ceph_health_status" {
					t.Fatal("lost metric query")
				}
				return jsonResponse(200, bodies[index]), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if _, err := client.Dashboards(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Query(ctx, "cluster_health", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Rules(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Alerts(ctx, "test-fsid"); err != nil {
				t.Fatal(err)
			}
			if _, err := client.Silences(ctx); err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			if _, err := client.CreateSilence(ctx, Silence{StartsAt: start, EndsAt: start.Add(time.Hour), Matchers: []Matcher{{Name: "alertname", Value: "CephHealth", IsEqual: true}}}); err != nil {
				t.Fatal(err)
			}
			if err := client.DeleteSilence(ctx, "silence id"); err != nil {
				t.Fatal(err)
			}
			if calls != len(paths) {
				t.Fatalf("calls %d", calls)
			}
		})
	}
}
