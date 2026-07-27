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
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

type externalRoundTripFunc func(*http.Request) (*http.Response, error)

func (f externalRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

const externalTestKey = "0123456789abcdefghijklmnopqrstuv"

func externalTestService(t *testing.T) (*Service, *endpointservice.Service, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: externalTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/external.db"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "test", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	endpoints := endpointservice.New(func() *store.Database { return db }, externalTestKey)
	return New(endpoints, nil), endpoints, cluster
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
			body = `{"status":"success","data":{"groups":[{"name":"ceph","rules":[{"name":"CephHealth","query":"ceph_health_status"}]}]}}`
		case request.URL.Host == "alertmanager.example.test" && request.URL.Path == "/api/v2/alerts":
			body = `[{"labels":{"alertname":"CephHealth"},"annotations":{},"status":{"state":"active"},"startsAt":"2026-07-26T00:00:00Z"}]`
		case request.URL.Host == "alertmanager.example.test" && request.URL.Path == "/api/v2/silences":
			body = `[]`
		case request.URL.Host == "grafana.example.test" && request.URL.Path == "/api/search":
			body = `[{"id":"1","uid":"ceph","title":"Ceph","url":"/d/ceph"}]`
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
	bucketID := base64.RawURLEncoding.EncodeToString([]byte("\x00bucket-one"))
	blockers, warnings, err := service.CheckPlan(ctx, operationservice.PlanRequest{ClusterID: cluster.ID, Action: "rgw_bucket.delete", ResourceKind: "rgw_bucket", ResourceKey: "rgw/bucket/" + bucketID})
	if err != nil || len(blockers) != 0 || len(warnings) != 0 {
		t.Fatalf("S3 plan pre-check blockers=%v warnings=%v err=%v", blockers, warnings, err)
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
