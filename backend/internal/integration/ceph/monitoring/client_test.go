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
