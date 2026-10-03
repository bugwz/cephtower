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

func TestBucketPolicyOriginalDocumentVerification(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	document := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"s3:GetObject","Resource":"arn:aws:s3:::bucket/*","Condition":{"NumericLessThan":{"s3:signatureAge":9007199254740993}}}]}`
	for _, tc := range []struct {
		name, body           string
		putStatus, getStatus int
		transportFailure     bool
		code                 string
	}{
		{"exact", document, 200, 200, false, ""},
		{"large integer changed", strings.Replace(document, "9007199254740993", "9007199254740992", 1), 200, 200, false, "post_check_failed"},
		{"permission changed", strings.Replace(document, "Allow", "Deny", 1), 200, 200, false, "post_check_failed"},
		{"format changed", " " + document, 200, 200, false, "post_check_failed"},
		{"invalid response", "broken", 200, 200, false, "post_check_failed"},
		{"empty response", "", 200, 200, false, "post_check_failed"},
		{"lost read access", "denied", 200, 403, false, "post_check_failed"},
		{"missing policy", `<Error><Code>NoSuchBucketPolicy</Code></Error>`, 200, 404, false, "post_check_failed"},
		{"read transport failure", "", 200, 200, true, "post_check_failed"},
		{"write failure", "", 503, 200, false, "s3_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != "/team:bucket" || r.URL.RawQuery != "policy=" {
					t.Fatal("wrong scoped target")
				}
				if calls == 1 {
					body, _ := io.ReadAll(r.Body)
					if r.Method != "PUT" || string(body) != document {
						t.Fatal("policy document changed before write")
					}
					return &http.Response{StatusCode: tc.putStatus, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
				}
				if calls != 2 || r.Method != "GET" {
					t.Fatal("unexpected request")
				}
				if tc.transportFailure {
					return nil, errors.New("read failed")
				}
				return &http.Response{StatusCode: tc.getStatus, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
			_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": "policy", "document": document}})
			wantCalls := 2
			if tc.putStatus != 200 {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("calls=%d expected=%d", calls, wantCalls)
			}
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Code != tc.code || failure.Retryable {
				t.Fatalf("unsafe error %v", err)
			}
		})
	}
}
