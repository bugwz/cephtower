package external

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBucketVersioningWriteVerification(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		input, status string
		put, get      int
		code          string
	}{
		{"enabled", "Enabled", 200, 200, ""}, {"suspended", "Suspended", 200, 200, ""},
		{"enabled", "Suspended", 200, 200, "post_check_failed"}, {"suspended", "", 200, 200, "post_check_failed"},
		{"enabled", "broken", 200, 200, "post_check_failed"}, {"enabled", "Enabled", 200, 403, "post_check_failed"},
		{"enabled", "Enabled", 503, 200, "s3_failed"}, {"disabled", "", 200, 200, "invalid_request"},
	} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Path != "/team:bucket" || r.URL.RawQuery != "versioning=" || r.Header.Get("Authorization") == "" {
				t.Fatalf("wrong target %s", r.URL)
			}
			status, body := tc.get, "<VersioningConfiguration/>"
			if tc.status != "" {
				body = "<VersioningConfiguration><Status>" + tc.status + "</Status></VersioningConfiguration>"
			}
			if calls == 1 {
				data, _ := io.ReadAll(r.Body)
				wanted := map[string]string{"enabled": "Enabled", "suspended": "Suspended"}[tc.input]
				if r.Method != "PUT" || !strings.Contains(string(data), "<Status>"+wanted+"</Status>") || strings.Contains(string(data), "MfaDelete") {
					t.Fatal("wrong write")
				}
				status, body = tc.put, ""
			} else if calls != 2 || r.Method != "GET" {
				t.Fatal("unexpected request")
			}
			return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.update", ResourceKey: "rgw/bucket/" + id, Parameters: map[string]any{"versioning": tc.input}})
		wantedCalls := 2
		if tc.code == "s3_failed" {
			wantedCalls = 1
		}
		if tc.code == "invalid_request" {
			wantedCalls = 0
		}
		if calls != wantedCalls {
			t.Fatalf("calls %d want %d", calls, wantedCalls)
		}
		if tc.code == "" {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != tc.code || failure.Retryable {
			t.Fatalf("unsafe failure %v", err)
		}
	}
}
