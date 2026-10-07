package external

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"cephtower/backend/internal/store"
)

type externalRoundTripFunc func(*http.Request) (*http.Response, error)

func (f externalRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

const externalTestKey = "0123456789abcdefghijklmnopqrstuv"

func externalTestService(t *testing.T) (*Service, *endpointservice.Service, store.CephCluster) {
	t.Helper()
	fsid := "00000000-0000-0000-0000-000000000001"
	return externalTestServiceWithFSID(t, &fsid)
}

func externalTestServiceWithFSID(t *testing.T, fsid *string) (*Service, *endpointservice.Service, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: externalTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "external.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "test", FSID: fsid, MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	endpoints := endpointservice.New(func() *store.Database { return db }, externalTestKey)
	return New(endpoints, "", nil), endpoints, cluster
}

func TestHTTPClientUsesEndpointTimeoutWithoutRequiringCredential(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	if _, err := endpoints.CreateEndpoint(context.Background(), cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test", TimeoutSeconds: 7}); err != nil {
		t.Fatal(err)
	}
	_, credential, client, err := service.httpClient(context.Background(), cluster.ID, "alertmanager")
	if err != nil {
		t.Fatal(err)
	}
	if credential.Token != "" || client.Timeout != 7*time.Second {
		t.Fatalf("credential=%#v timeout=%s", credential, client.Timeout)
	}
}

func TestMetricRangeFailuresArePropagated(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "prometheus", URL: "https://prometheus.example.test"}); err != nil {
		t.Fatal(err)
	}
	query := url.Values{"metric_id": {"pool_read_bytes"}, "start": {"2026-01-01T00:00:00Z"}, "end": {"2026-01-01T01:00:00Z"}, "step": {"30s"}}
	for _, tc := range []struct {
		status int
		body   string
		valid  bool
	}{
		{503, `unavailable`, false},
		{200, `{"status":"error","data":{"resultType":"matrix","result":[]}}`, false},
		{200, `{}`, false},
		{200, `{"status":"success","data":{"resultType":"vector","result":[]}}`, false},
		{200, `{"status":"success","data":{"resultType":"matrix","result":null}}`, false},
		{200, `{"status":"success","data":{"resultType":"matrix","result":[]}}`, true},
	} {
		s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/api/v1/query_range" {
				t.Fatalf("path %s", r.URL.Path)
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body)), Request: r}, nil
		})
		result, err := s.readMetric(ctx, cluster.ID, "metric/range", query)
		if tc.valid {
			if err != nil || result == nil {
				t.Fatalf("valid empty result: %v", err)
			}
			meta := result.(map[string]any)["meta"].(map[string]any)
			if meta["start"] != "2026-01-01T00:00:00Z" || meta["end"] != "2026-01-01T01:00:00Z" || meta["step_seconds"] != float64(30) {
				t.Fatalf("missing evaluated range: %#v", meta)
			}
		} else if err == nil || result != nil {
			t.Fatalf("failure lost for %s: %#v %v", tc.body, result, err)
		}
	}
}

func TestMetricNotices(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "prometheus", URL: "https://prometheus.test"}); err != nil {
		t.Fatal(err)
	}
	for _, history := range []bool{false, true} {
		kind, key := "vector", "metric/query"
		query := url.Values{"metric_id": {"smb_metrics_status"}}
		if history {
			kind, key = "matrix", "metric/range"
			query.Set("start", "2026-01-01T00:00:00Z")
			query.Set("end", "2026-01-01T01:00:00Z")
			query.Set("step", "30s")
		}
		s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			body := `{"status":"success","warnings":["partial data"],"infos":["sample omitted"],"data":{"resultType":"` + kind + `","result":[]}}`
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		result, err := s.readMetric(ctx, cluster.ID, key, query)
		if err != nil {
			t.Fatal(err)
		}
		meta := result.(map[string]any)["meta"].(map[string]any)
		if meta["warnings"].([]string)[0] != "partial data" || meta["infos"].([]string)[0] != "sample omitted" {
			t.Fatal("lost query notices")
		}
	}
}

func TestHTTPClientRejectsMalformedConfiguredCredential(t *testing.T) {
	_, endpoints, cluster := externalTestService(t)
	if _, err := endpoints.CreateEndpoint(context.Background(), cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(context.Background(), cluster.ID, endpointservice.CredentialInput{Kind: "alertmanager", Value: map[string]any{"unexpected": "secret"}}); err == nil {
		t.Fatal("malformed credential was accepted")
	}
}

func TestProtocolNativeHTTPReadsUseTypedAdapters(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	for _, endpoint := range []endpointservice.EndpointInput{
		{Kind: "prometheus", URL: "https://prometheus.example.test"},
		{Kind: "alertmanager", URL: "https://alertmanager.example.test"},
		{Kind: "grafana", URL: "https://grafana.example.test"},
		{Kind: "iscsi", URL: "https://iscsi.example.test"},
		{Kind: "s3", URL: "https://s3.example.test"},
	} {
		if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpoint); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "iscsi", Value: map[string]any{"username": "gateway-user", "password": "gateway-password"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret", "region": "us-east-1"}}); err != nil {
		t.Fatal(err)
	}
	service.transport = externalRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "{}"
		switch {
		case request.URL.Host == "prometheus.example.test" && request.URL.Path == "/api/v1/query":
			body = `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"job":"ceph"},"value":[1,"1"]}]}}`
		case request.URL.Host == "prometheus.example.test" && request.URL.Path == "/api/v1/rules":
			body = `{"status":"success","data":{"groups":[{"name":"ceph","rules":[{"type":"alerting","name":"CephHealth","query":"ceph_health_status"}]}]}}`
		case request.URL.Host == "alertmanager.example.test" && request.URL.Path == "/api/v2/alerts":
			if request.URL.Query().Get("filter") != `cluster="00000000-0000-0000-0000-000000000001"` {
				t.Fatalf("missing cluster filter: %s", request.URL)
			}
			body = `[{"labels":{"alertname":"CephHealth"},"annotations":{},"status":{"state":"active"},"startsAt":"2026-07-26T00:00:00Z"}]`
		case request.URL.Host == "alertmanager.example.test" && request.URL.Path == "/api/v2/alerts/groups":
			if request.URL.Query().Get("filter") != `cluster="00000000-0000-0000-0000-000000000001"` {
				t.Fatalf("missing cluster filter: %s", request.URL)
			}
			body = `[{"labels":{"cluster":"ceph"},"receiver":{"name":"email"},"alerts":[]}]`
		case request.URL.Host == "alertmanager.example.test" && request.URL.Path == "/api/v2/silences":
			body = `[]`
		case request.URL.Host == "grafana.example.test" && request.URL.Path == "/api/search":
			body = `[{"id":1,"uid":"ceph","title":"Ceph","url":"/d/ceph"}]`
		case request.URL.Host == "iscsi.example.test" && request.URL.Path == "/api/gateway":
			body = `{"status":"ok"}`
		case request.URL.Host == "iscsi.example.test" && request.URL.Path == "/api/target":
			body = `[{"iqn":"iqn.2026-07.test","portals":[],"disks":[],"clients":[],"groups":[]}]`
		case request.URL.Host == "s3.example.test" && request.Method == http.MethodGet && request.URL.Query().Has("policy"):
			body = `{"Version":"2012-10-17","Statement":[]}`
		case request.URL.Host == "s3.example.test" && request.Method == http.MethodHead:
			body = ""
		default:
			t.Fatalf("unexpected request %s %s", request.Method, request.URL.String())
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": []string{"application/json"}}, Request: request}, nil
	})

	checks := []struct {
		kind, key string
		query     url.Values
		contains  string
	}{
		{"metric", "metric/query", url.Values{"metric_id": []string{"cluster_health"}}, "result_type"},
		{"alert", "", nil, "CephHealth"},
		{"alert_group", "", nil, "email"},
		{"alert_rule", "", nil, "ceph_health_status"},
		{"silence", "", nil, "items"},
		{"grafana", "", nil, "Ceph"},
		{"iscsi_gateway", "", nil, "ok"},
		{"iscsi_target", "", nil, "iqn.2026-07.test"},
		{"rgw_bucket_policy", base64.RawURLEncoding.EncodeToString([]byte("\x00bucket-one")), nil, "2012-10-17"},
	}
	for _, check := range checks {
		result, err := service.Read(ctx, cluster.ID, check.kind, check.key, check.query)
		if err != nil {
			t.Fatalf("Read(%s): %v", check.kind, err)
		}
		if !strings.Contains(toJSON(t, result), check.contains) {
			t.Fatalf("Read(%s) = %#v", check.kind, result)
		}
	}
}

func toJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func TestAlertRulesFlattenGroupsAndPreserveDetails(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "prometheus", URL: "https://prometheus.example.test"}); err != nil {
		t.Fatal(err)
	}
	const body = `{"status":"success","data":{"groups":[{"name":"ceph","file":"a.yml","rules":[{"type":"recording","name":"record","query":"sum(up)"},{"type":"alerting","name":"Health","query":"ceph_health_status > 0","duration":0,"labels":{"severity":"warning"},"annotations":{"summary":"bad health"},"health":"ok","state":"firing","alerts":[{"value":"9007199254740993","state":"firing"}],"lastError":"evaluation failed","lastEvaluation":"2026-01-01T00:00:00Z","evaluationTime":0}]},{"name":"ceph","file":"b.yml","rules":[{"type":"alerting","name":"Health","query":"up == 0"}]}]}}`
	s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/rules" {
			t.Fatalf("unexpected path: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	result, err := s.Read(ctx, cluster.ID, "alert_rule", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	rows := result.(map[string]any)["items"].([]alertRuleRow)
	if len(rows) != 2 || rows[0].Group != "ceph" || rows[0].File != "a.yml" || rows[1].File != "b.yml" || rows[0].RuleKey == rows[1].RuleKey {
		t.Fatalf("flattened rules: %#v", rows)
	}
	first := rows[0]
	if first.Name != "Health" || first.State != "firing" || first.Health != "ok" || first.Duration == nil || *first.Duration != 0 || first.EvaluationTime == nil || *first.EvaluationTime != 0 || first.LastEvaluation == nil || first.LastError != "evaluation failed" {
		t.Fatalf("lost rule fields: %#v", first)
	}
	if !strings.Contains(string(first.Alerts), "9007199254740993") || first.Annotations["summary"] != "bad health" || first.Labels["severity"] != "warning" {
		t.Fatalf("lost details: %#v", first)
	}
}

func TestSilenceExpiryUsesCompleteTarget(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test"}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v2/silence/silence-a" {
			t.Fatalf("unexpected target: %s %s", r.Method, r.URL)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})
	for _, key := range []string{"silence-a", "other/silence-a", "silence/", "silence/parent/silence-a", "silence/ silence-a", "silence/silence-a ", "silence/.", "silence/..", "silence/a\nb"} {
		if _, err := s.Execute(ctx, Request{ClusterID: cluster.ID, Action: "silence.delete", ResourceKey: key}); err == nil {
			t.Fatalf("accepted %q", key)
		}
	}
	if calls != 0 {
		t.Fatalf("invalid target sent %d requests", calls)
	}
	if _, err := s.Execute(ctx, Request{ClusterID: cluster.ID, Action: "silence.delete", ResourceKey: "silence/silence-a"}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected one expiry, got %d", calls)
	}
}

func TestAlertmanagerListsRejectAbsentAndNullData(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test"}); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"alert", "silence"} {
		for _, body := range []string{"", "null", "{}", "[null]", "[{},null]", "[]"} {
			s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			result, err := s.Read(ctx, cluster.ID, kind, "", nil)
			if body == "[]" {
				if err != nil || !strings.Contains(toJSON(t, result), `"items":[]`) {
					t.Fatalf("valid empty %s: %#v %v", kind, result, err)
				}
			} else if err == nil || result != nil {
				t.Fatalf("accepted unavailable %s %q: %#v %v", kind, body, result, err)
			}
		}
	}
}
