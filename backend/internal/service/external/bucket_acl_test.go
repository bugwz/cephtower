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

func TestBucketACLReadIsSignedAndScoped(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	document := `<AccessControlPolicy><Owner><ID>owner</ID></Owner><AccessControlList/></AccessControlPolicy>`
	for _, tc := range []struct {
		body   string
		status int
		valid  bool
	}{
		{document, 200, true}, {"broken", 200, false}, {"denied", 403, false},
		{`<Error><Code>NoSuchBucket</Code></Error>`, 404, false},
		{`<Error><Code>NoSuchKey</Code></Error>`, 404, false},
	} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.Method != "GET" || r.URL.Path != "/team:bucket" || r.URL.RawQuery != "acl=" || r.Header.Get("Authorization") == "" {
				t.Fatal("wrong signed tenant-scoped request")
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"acl"}})
		if calls != 1 || (err == nil) != tc.valid {
			t.Fatalf("unexpected result: %v, calls %d", err, calls)
		}
		if !tc.valid {
			continue
		}
		row := result.(map[string]any)
		acl := row["acl"].(s3.BucketACLConfiguration)
		if row["bucket_id"] != id || row["configured"] != true || row["document"] != document || acl.Owner.ID != "owner" || acl.Grants == nil {
			t.Fatal("lost native ACL data")
		}
	}
}
