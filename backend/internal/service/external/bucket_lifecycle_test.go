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

func TestBucketLifecycleRead(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	document := `<LifecycleConfiguration><Rule><ID>rule</ID><Status>Enabled</Status><Filter/><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>`
	for _, tc := range []struct {
		body              string
		status            int
		valid, configured bool
	}{
		{document, 200, true, true}, {`<Error><Code>NoSuchLifecycleConfiguration</Code></Error>`, 404, true, false},
		{`<LifecycleConfiguration/>`, 200, false, false}, {`<Error><Code>NoSuchBucket</Code></Error>`, 404, false, false},
		{`denied`, 403, false, false},
	} {
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/team:bucket" || r.URL.RawQuery != "lifecycle=" {
				t.Fatal("wrong scoped read")
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"lifecycle"}})
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
		rules := row["lifecycle_rules"].([]s3.BucketLifecycleRule)
		if row["configured"] != tc.configured {
			t.Fatal("wrong configuration state")
		}
		if tc.configured {
			if len(rules) != 1 || rules[0].ID != "rule" || row["document"] != document {
				t.Fatal("lost rule or raw document")
			}
		} else if len(rules) != 0 || rules == nil || row["document"] != nil {
			t.Fatal("missing configuration misrepresented")
		}
	}
}
