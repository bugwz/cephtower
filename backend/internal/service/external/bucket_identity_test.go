package external

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestBucketCreationUsesExplicitTenantIdentity(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tenant := range []string{"", "team", "other"} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			method := "PUT"
			if calls == 2 {
				method = "HEAD"
			}
			if r.Method != method || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != "" || r.Header.Get("Authorization") == "" {
				t.Fatalf("wrong create target %s", r.URL)
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})
		result, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.create", Parameters: map[string]any{"name": "bucket", "tenant": tenant}})
		id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
		if err != nil || calls != 2 || result.ResourceURL != fmt.Sprintf("/api/v1/cluster/%d/rgw/bucket/%s", cluster.ID, id) {
			t.Fatalf("wrong identity %+v %v", result, err)
		}
	}
	service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid identity reached network")
		return nil, nil
	})
	for _, input := range []map[string]any{{"name": ""}, {"name": "team:bucket"}, {"name": "../bucket"}, {"name": "bucket", "tenant": "a/b"}, {"name": "bucket", "tenant": true}, {"name": "bucket", "tenant": "a\x00b"}, {"bucket": "legacy"}} {
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.create", Parameters: input})
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "invalid_request" || failure.Retryable {
			t.Fatalf("unsafe failure %v", err)
		}
	}
}

func TestBucketRequestsPreserveFullInventoryIdentity(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tenant := range []string{"", "team", "other"} {
		id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00same-bucket"))
		for _, tc := range []struct{ action, method, kind, document string }{
			{"rgw_bucket_policy.update", "PUT", "policy", "{}"},
			{"rgw_bucket_policy.update", "PUT", "cors", "<CORSConfiguration><CORSRule><AllowedOrigin>*</AllowedOrigin></CORSRule></CORSConfiguration>"},
			{"rgw_bucket_policy.update", "PUT", "lifecycle", "<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter/><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>"},
			{"rgw_bucket_policy.update", "PUT", "encryption", "<ServerSideEncryptionConfiguration/>"},
			{"rgw_bucket.update", "PUT", "versioning", ""},
			{"rgw_bucket.delete", "DELETE", "", ""},
		} {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				query := ""
				if tc.kind != "" {
					query = tc.kind + "="
				}
				method, body := tc.method, ""
				if (tc.kind == "encryption" || tc.kind == "cors" || tc.kind == "policy") && calls == 2 {
					method, body = "GET", tc.document
				}
				if tc.kind == "lifecycle" && calls == 2 {
					method, body = "GET", strings.Replace(tc.document, "<Rule>", "<Rule><ID>generated</ID>", 1)
				}
				if tc.kind == "versioning" && calls == 2 {
					method, body = "GET", "<VersioningConfiguration><Status>Enabled</Status><MfaDelete>Disabled</MfaDelete></VersioningConfiguration>"
				}
				if r.Method != method || r.URL.Path != "/"+tenant+":same-bucket" || r.URL.RawQuery != query || r.Header.Get("Authorization") == "" {
					t.Fatalf("wrong target: %s %s", r.Method, r.URL)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			key := "rgw/bucket/" + id
			if tc.action == "rgw_bucket_policy.update" {
				key += "/policy"
			}
			result, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: tc.action, ResourceKey: key, Parameters: map[string]any{"kind": tc.kind, "document": tc.document, "versioning": "enabled"}})
			wantCalls := 1
			if tc.kind == "encryption" || tc.kind == "versioning" || tc.kind == "cors" || tc.kind == "lifecycle" || tc.kind == "policy" {
				wantCalls = 2
			}
			if err != nil || calls != wantCalls || result.ResourceURL != fmt.Sprintf("/api/v1/cluster/%d/rgw/bucket/%s", cluster.ID, id) {
				t.Fatalf("identity lost: %+v %v calls=%d", result, err, calls)
			}
		}
		for _, kind := range []string{"policy", "cors", "lifecycle", "encryption"} {
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != "GET" || r.URL.Path != "/"+tenant+":same-bucket" || r.URL.RawQuery != kind+"=" {
					t.Fatalf("wrong read target: %s", r.URL)
				}
				body := "native"
				if kind == "lifecycle" {
					body = "<LifecycleConfiguration><Rule><Status>Enabled</Status><Filter/><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>"
				}
				if kind == "cors" {
					body = "<CORSConfiguration><CORSRule><AllowedOrigin>*</AllowedOrigin></CORSRule></CORSConfiguration>"
				}
				if kind == "encryption" {
					body = "<ServerSideEncryptionConfiguration/>"
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {kind}})
			if err != nil || result.(map[string]any)["bucket_id"] != id {
				t.Fatalf("read identity lost: %v %v", result, err)
			}
		}
	}
	service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid identity reached network")
		return nil, nil
	})
	for _, key := range []string{"rgw/bucket/AGJ1Y2tldA", "rgw/bucket/AGJ1Y2tldA/policy/extra", "rgw/user/AGJ1Y2tldA/policy", "rgw/bucket/AGJ1Y2tldA/policy/", "rgw/bucket/invalid/policy"} {
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: key, Parameters: map[string]any{"kind": "policy", "document": "{}"}})
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "invalid_request" || failure.Retryable {
			t.Fatalf("unsafe error for %s: %v", key, err)
		}
	}
}

func TestBucketIdentityRejectsAmbiguousEncodings(t *testing.T) {
	for _, raw := range []string{"bucket", "\x00", "tenant\x00bucket\x00extra", "tenant:other\x00bucket", "tenant\x00other:bucket", "team/other\x00bucket", "\x00bucket/other", "team\n\x00bucket", "\x00 bucket", "\xff\x00bucket", "\x00bucket\\other"} {
		if _, err := decodeBucketID(base64.RawURLEncoding.EncodeToString([]byte(raw))); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	for _, id := range []string{"AGJ1Y2tldA==", "AGJ1Y2tldB", "AGJ1\nY2tldA", " AGJ1Y2tldA"} {
		if _, err := decodeBucketID(id); err == nil {
			t.Fatalf("accepted %q", id)
		}
	}
}
