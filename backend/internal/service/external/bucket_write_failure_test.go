package external

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestBucketCreateDeleteFailuresAreNotRetried(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, status := range []int{0, 403, 404, 503} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Path != "/team:bucket" || r.Header.Get("Authorization") == "" {
				t.Fatal("wrong verification identity")
			}
			code := 200
			if calls == 1 {
				if r.Method != "PUT" {
					t.Fatal("expected create")
				}
			} else {
				if calls != 2 || r.Method != "HEAD" {
					t.Fatal("expected single verification")
				}
				if status == 0 {
					return nil, context.DeadlineExceeded
				}
				code = status
			}
			return &http.Response{StatusCode: code, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.create", Parameters: map[string]any{"name": "bucket", "tenant": "team"}})
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable || calls != 2 {
			t.Fatalf("unsafe verification result %v calls=%d", err, calls)
		}
	}
	for _, action := range []string{"rgw_bucket.create", "rgw_bucket.delete"} {
		for _, status := range []int{0, 403, 404, 409, 500, 503} {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				method := "PUT"
				if action == "rgw_bucket.delete" {
					method = "DELETE"
				}
				if r.Method != method || r.URL.Path != "/team:bucket" {
					t.Fatalf("wrong request %s %s", r.Method, r.URL)
				}
				if status == 0 {
					return nil, context.DeadlineExceeded
				}
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("<Error><Code>RequestFailed</Code></Error>"))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
			_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: action, ResourceKey: "rgw/bucket/" + id, Parameters: map[string]any{"name": "bucket", "tenant": "team"}})
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Code != "s3_failed" || failure.Retryable || !strings.Contains(failure.Message, "refresh before another change") || calls != 1 {
				t.Fatalf("%s status=%d calls=%d unsafe error %v", action, status, calls, err)
			}
		}
	}
}
