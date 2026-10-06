package monitoring

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPrometheusUsesRegisteredQueryAndBearerToken(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/query" || r.URL.Query().Get("query") != "ceph_health_status" {
			t.Fatalf("unexpected request %s", r.URL.String())
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatal("bearer token missing")
		}
		return jsonResponse(200, `{"status":"success","data":{"resultType":"vector","result":[]}}`), nil
	})}
	client, err := New("https://prometheus.example.test", "token", httpClient)
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Query(context.Background(), "cluster_health", nil)
	if err != nil || result.Status != "success" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if _, err := client.Query(context.Background(), "arbitrary_promql", nil); err == nil {
		t.Fatal("unregistered query accepted")
	}
}

func TestAlertmanagerReadPreservesDashboardFields(t *testing.T) {
	const body = `[{"labels":{"alertname":"CephHealth","severity":"warning"},"annotations":{"summary":"health warning","description":"details","impact":"degraded","fix":"inspect"},"status":{"state":"suppressed","silencedBy":["silence-1"],"inhibitedBy":["other-alert"]},"startsAt":"2026-01-01T00:00:00Z","endsAt":"2026-01-01T01:00:00Z","updatedAt":"2026-01-01T00:30:00Z","fingerprint":"001abc","generatorURL":"https://prometheus.example.test/graph?g0.expr=ceph_health_status","receivers":[{"name":"ceph"}]}]`
	client, err := New("https://alertmanager.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/alerts" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
		return jsonResponse(200, body), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	alerts, err := client.Alerts(context.Background(), "test-fsid")
	if err != nil || len(alerts) != 1 {
		t.Fatalf("alerts=%#v err=%v", alerts, err)
	}
	encoded, err := json.Marshal(alerts)
	if err != nil {
		t.Fatal(err)
	}
	var want, got any
	if err := json.Unmarshal([]byte(body), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dashboard fields lost: %s", encoded)
	}
}

func TestSilenceReadPreservesStateWithoutChangingCreatePayload(t *testing.T) {
	for _, state := range []string{"active", "pending", "expired", "future-state"} {
		body := `[{"id":"silence-1","matchers":[{"name":"alertname","value":"CephHealth","isRegex":false,"isEqual":true}],"startsAt":"2026-01-01T00:00:00Z","endsAt":"2026-01-01T02:00:00Z","createdBy":"operator","comment":"maintenance","status":{"state":"` + state + `"},"updatedAt":"2026-01-01T00:30:00Z"}]`
		client, err := New("https://alertmanager.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet || r.URL.Path != "/api/v2/silences" {
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
			}
			return jsonResponse(200, body), nil
		})})
		if err != nil {
			t.Fatal(err)
		}
		rows, err := client.Silences(context.Background())
		if err != nil || len(rows) != 1 {
			t.Fatalf("rows=%#v err=%v", rows, err)
		}
		if rows[0].Status == nil || rows[0].Status.State != state || rows[0].UpdatedAt == nil || rows[0].UpdatedAt.Format(time.RFC3339) != "2026-01-01T00:30:00Z" {
			t.Fatalf("lost metadata: %#v", rows[0])
		}
		encoded, err := json.Marshal(rows[0].Silence)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), "status") || strings.Contains(string(encoded), "updatedAt") {
			t.Fatalf("read fields leaked to write payload: %s", encoded)
		}
	}
}

func TestHostMetricQueriesAreRegistered(t *testing.T) {
	for _, metricID := range []string{
		"host_cpu_usage",
		"host_memory_usage",
		"host_disk_read_bytes",
		"host_disk_write_bytes",
		"host_network_receive",
		"host_network_transmit",
	} {
		if strings.TrimSpace(metricQueries[metricID]) == "" {
			t.Fatalf("host metric %q is not registered", metricID)
		}
	}
}

func TestRulesRejectUnavailableGroups(t *testing.T) {
	for _, body := range []string{`{}`, `{"status":"error","data":{"groups":[]}}`, `{"status":"success","data":{"groups":null}}`, `{"status":"success","data":{"groups":[{"name":"a"}]}}`} {
		client, err := New("https://prometheus.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { return jsonResponse(200, body), nil })})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Rules(context.Background()); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
}

func TestInstantMetricRejectsInvalidEnvelope(t *testing.T) {
	for _, body := range []string{`{}`, `{"status":"error"}`, `{"status":"success","data":{"resultType":"matrix","result":[]}}`, `{"status":"success","data":{"resultType":"vector","result":null}}`} {
		client, err := New("https://prometheus.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { return jsonResponse(200, body), nil })})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Query(context.Background(), "pool_read_bytes", nil); err == nil {
			t.Fatalf("accepted %s", body)
		}
	}
}

func TestPoolMetricRangeRetainsPoolIdentity(t *testing.T) {
	for metric, counter := range map[string]string{"pool_read_bytes": "rd_bytes", "pool_write_bytes": "wr_bytes", "pool_read_ops": "rd", "pool_write_ops": "wr"} {
		t.Run(metric, func(t *testing.T) {
			client, err := New("https://prometheus.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				want := "sum by (pool_id) (rate(ceph_pool_" + counter + "[5m]))"
				if r.URL.Path != "/api/v1/query_range" || r.URL.Query().Get("query") != want || r.URL.Query().Get("step") != "30" {
					t.Fatalf("request = %s", r.URL)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"pool_id":"7"},"values":[[1000,"1.5"]]}]}}`), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.QueryRange(context.Background(), metric, time.Unix(1000, 0), time.Unix(4600, 0), 30*time.Second)
			if err != nil || result.Data.ResultType != "matrix" || len(result.Data.Result) != 1 {
				t.Fatalf("result=%#v err=%v", result, err)
			}
		})
	}
}

func TestAlertmanagerSilenceUsesTypedPayload(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v2/silences" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var value map[string]any
		if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
			t.Fatal(err)
		}
		if _, ok := value["matchers"]; !ok {
			t.Fatal("matchers missing")
		}
		return jsonResponse(200, `{"silenceID":"silence-1"}`), nil
	})}
	client, _ := New("https://alertmanager.example.test", "", httpClient)
	now := time.Now().UTC()
	id, err := client.CreateSilence(context.Background(), Silence{Matchers: []Matcher{{Name: "alertname", Value: "CephHealth", IsEqual: true}}, StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: "test", Comment: "test"})
	if err != nil || id != "silence-1" {
		t.Fatalf("id=%q err=%v", id, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestClientRejectsEmbeddedCredentials(t *testing.T) {
	if _, err := New("https://user:password@example.test", "", nil); err == nil || !strings.Contains(err.Error(), "without embedded") {
		t.Fatalf("unexpected error %v", err)
	}
}
