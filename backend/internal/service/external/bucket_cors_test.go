package external

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBucketCORSWriteVerification(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	sent := `<CORSConfiguration><CORSRule><AllowedOrigin>*</AllowedOrigin><AllowedMethod>get</AllowedMethod><ExposeHeader>a</ExposeHeader><ExposeHeader>b</ExposeHeader></CORSRule></CORSConfiguration>`
	for _, tc := range []struct {
		body              string
		status            int
		valid, configured bool
	}{
		{sent, 200, true, true}, {`<Error><Code>NoSuchCORSConfiguration</Code></Error>`, 404, true, false}, {`<CORSConfiguration/>`, 200, false, false}, {`<Error><Code>NoSuchBucket</Code></Error>`, 404, false, false},
	} {
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/team:bucket" || r.URL.RawQuery != "cors=" {
				t.Fatal("wrong read")
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"cors"}})
		if !tc.valid {
			if err == nil {
				t.Fatal("invalid response accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		row := result.(map[string]any)
		rules := row["cors_rules"].([]s3.BucketCORSRule)
		if row["configured"] != tc.configured {
			t.Fatal("wrong configured state")
		}
		if tc.configured {
			if len(rules) != 1 || rules[0].AllowedMethods[0] != "GET" || row["document"] != sent {
				t.Fatal("lost rule data")
			}
		} else if len(rules) != 0 || row["document"] != nil {
			t.Fatal("missing configuration misrepresented")
		}
	}
	for _, tc := range []struct {
		body   string
		status int
		valid  bool
	}{
		{strings.Replace(sent, "get", "GET", 1), 200, true},
		{strings.Replace(sent, "get", "PUT", 1), 200, false},
		{strings.Replace(sent, "<ExposeHeader>a</ExposeHeader><ExposeHeader>b</ExposeHeader>", "<ExposeHeader>b</ExposeHeader><ExposeHeader>a</ExposeHeader>", 1), 200, false},
		{"broken", 200, false}, {sent, 403, false}, {`<Error><Code>NoSuchCORSConfiguration</Code></Error>`, 404, false},
	} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Path != "/team:bucket" || r.URL.RawQuery != "cors=" {
				t.Fatal("wrong scoped target")
			}
			status, body := tc.status, tc.body
			if calls == 1 {
				if r.Method != "PUT" {
					t.Fatal("expected write")
				}
				status, body = 200, ""
			} else if calls != 2 || r.Method != "GET" {
				t.Fatal("expected verification")
			}
			return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": "cors", "document": sent}})
		if calls != 2 {
			t.Fatalf("calls=%d", calls)
		}
		if tc.valid {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatalf("unsafe error %v", err)
		}
	}
}
