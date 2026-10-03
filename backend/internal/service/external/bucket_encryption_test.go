package external

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/s3"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

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
