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
	"cephtower/backend/internal/integration/ceph/s3"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestBucketEncryptionWriteVerification(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	wrap := func(inner string) string {
		return "<ServerSideEncryptionConfiguration>" + inner + "</ServerSideEncryptionConfiguration>"
	}
	empty := wrap("")
	rule := wrap("<Rule/>")
	defaults := wrap("<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm/><KMSMasterKeyID/></ApplyServerSideEncryptionByDefault><BucketKeyEnabled>false</BucketKeyEnabled></Rule>")
	kms := wrap("<Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>aws:kms</SSEAlgorithm><KMSMasterKeyID> key </KMSMasterKeyID></ApplyServerSideEncryptionByDefault><BucketKeyEnabled>2</BucketKeyEnabled></Rule>")
	for _, tc := range []struct {
		name, sent, returned string
		putStatus, getStatus int
		code                 string
	}{
		{"empty", empty, empty, 200, 200, ""},
		{"native defaults", defaults, rule, 200, 200, ""},
		{"normalized bool", kms, strings.Replace(kms, ">2<", ">true<", 1), 200, 200, ""},
		{"rule missing", rule, empty, 200, 200, "post_check_failed"},
		{"algorithm changed", kms, strings.Replace(kms, "aws:kms", "AES256", 1), 200, 200, "post_check_failed"},
		{"key changed", kms, strings.Replace(kms, " key ", "key", 1), 200, 200, "post_check_failed"},
		{"bucket key changed", kms, strings.Replace(kms, ">2<", ">0<", 1), 200, 200, "post_check_failed"},
		{"broken response", kms, "broken", 200, 200, "post_check_failed"},
		{"missing response", kms, `<Error><Code>ServerSideEncryptionConfigurationNotFoundError</Code></Error>`, 200, 404, "post_check_failed"},
		{"forbidden", kms, "", 200, 403, "post_check_failed"},
		{"write failure", kms, kms, 503, 200, "s3_failed"},
		{"invalid input", "broken", kms, 200, 200, "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != "/team:bucket" || r.URL.RawQuery != "encryption=" || r.Header.Get("Authorization") == "" {
					t.Fatalf("wrong target %s", r.URL)
				}
				status, body := tc.getStatus, tc.returned
				if calls == 1 {
					data, _ := io.ReadAll(r.Body)
					if r.Method != "PUT" || string(data) != tc.sent || r.Header.Get("Content-Type") != "application/xml" || r.Header.Get("Content-MD5") == "" {
						t.Fatal("wrong write")
					}
					status, body = tc.putStatus, ""
				} else if calls != 2 || r.Method != "GET" {
					t.Fatal("unexpected request")
				}
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
			_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": "encryption", "document": tc.sent}})
			wantCalls := 2
			if tc.code == "invalid_request" {
				wantCalls = 0
			} else if tc.code == "s3_failed" {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("calls=%d want=%d", calls, wantCalls)
			}
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Code != tc.code || failure.Retryable {
				t.Fatalf("unsafe failure %v", err)
			}
		})
	}
}

func TestBucketEncryptionStructuredRead(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		body              string
		status            int
		valid, configured bool
		want              s3.BucketEncryptionConfiguration
	}{
		{`<ServerSideEncryptionConfiguration/>`, 200, true, true, s3.BucketEncryptionConfiguration{}},
		{`<ServerSideEncryptionConfiguration><Rule/></ServerSideEncryptionConfiguration>`, 200, true, true, s3.BucketEncryptionConfiguration{RuleExists: true}},
		{`<ServerSideEncryptionConfiguration><Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>aws:kms</SSEAlgorithm><KMSMasterKeyID> key&amp;id </KMSMasterKeyID></ApplyServerSideEncryptionByDefault><BucketKeyEnabled>2</BucketKeyEnabled></Rule></ServerSideEncryptionConfiguration>`, 200, true, true, s3.BucketEncryptionConfiguration{RuleExists: true, Algorithm: "aws:kms", KMSMasterKeyID: " key&id ", BucketKeyEnabled: true}},
		{`<Error><Code>ServerSideEncryptionConfigurationNotFoundError</Code></Error>`, 404, true, false, s3.BucketEncryptionConfiguration{}},
		{`<Error><Code>NoSuchBucket</Code></Error>`, 404, false, false, s3.BucketEncryptionConfiguration{}},
		{`<Error><Code>AccessDenied</Code></Error>`, 403, false, false, s3.BucketEncryptionConfiguration{}},
		{`<ServerSideEncryptionConfiguration><Unknown/></ServerSideEncryptionConfiguration>`, 200, false, false, s3.BucketEncryptionConfiguration{}},
		{`broken`, 200, false, false, s3.BucketEncryptionConfiguration{}},
	} {
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/team:bucket" || r.URL.RawQuery != "encryption=" || r.Header.Get("Authorization") == "" {
				t.Fatalf("wrong request %s", r.URL)
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": {"application/xml"}}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"encryption"}})
		if !tc.valid {
			if err == nil {
				t.Fatalf("accepted invalid response %s", tc.body)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		row := result.(map[string]any)
		if row["configured"] != tc.configured || row["bucket_id"] != id {
			t.Fatalf("wrong state %+v", row)
		}
		if tc.configured {
			if row["encryption"] != tc.want || row["document"] != tc.body {
				t.Fatalf("lost metadata %+v", row)
			}
		} else if value, ok := row["encryption"]; !ok || value != nil || row["document"] != nil {
			t.Fatalf("missing state %+v", row)
		}
	}
}
