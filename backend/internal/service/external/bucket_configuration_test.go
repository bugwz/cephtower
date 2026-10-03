package external

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestBucketConfigurationWritesUseExplicitDocument(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		kind     string
		document any
		status   int
		wantCode string
		calls    int
	}{
		{"cors", "<CORSConfiguration/>", 200, "", 1}, {"policy", `{"Statement":[]}`, 200, "", 1}, {"cors", map[string]any{}, 200, "invalid_request", 0}, {"encryption", "{}", 200, "invalid_request", 0}, {"lifecycle", "<LifecycleConfiguration/>", 503, "s3_failed", 1},
	} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			body, _ := io.ReadAll(r.Body)
			if string(body) != tc.document || r.URL.Query().Has(tc.kind) == false {
				t.Fatalf("wrong request %s %s", r.URL, body)
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: "rgw/bucket/" + base64.RawURLEncoding.EncodeToString([]byte("\x00bucket")), Parameters: map[string]any{"kind": tc.kind, "document": tc.document}})
		if calls != tc.calls {
			t.Fatalf("calls=%d want=%d", calls, tc.calls)
		}
		if tc.wantCode == "" {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != tc.wantCode || failure.Retryable {
			t.Fatalf("unsafe error: %v", err)
		}
	}
	for _, kind := range []string{"policy", "cors", "lifecycle", "encryption"} {
		body := "native-" + kind
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/bucket" || r.URL.RawQuery != kind+"=" {
				t.Fatalf("wrong scoped read: %s %s", r.Method, r.URL)
			}
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/xml"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		result, err := service.readBucketPolicy(ctx, cluster.ID, base64.RawURLEncoding.EncodeToString([]byte("\x00bucket")), url.Values{"kind": []string{kind}})
		if err != nil {
			t.Fatal(err)
		}
		row := result.(map[string]any)
		if row["kind"] != kind || row["document"] != body || row["content_type"] != "application/xml" {
			t.Fatalf("read document lost: %+v", row)
		}
	}
}
