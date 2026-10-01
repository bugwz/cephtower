package monitoring

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
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
