package external

import (
	"cephtower/backend/internal/integration/ceph/s3"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestBucketObjectLockRead(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	document := `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>COMPLIANCE</Mode><Years>2</Years></DefaultRetention></Rule></ObjectLockConfiguration>`
	for _, tc := range []struct {
		body              string
		status            int
		valid, configured bool
	}{
		{document, 200, true, true}, {`<Error><Code>ObjectLockConfigurationNotFoundError</Code></Error>`, 404, true, false},
		{`<Error><Code>NoSuchBucket</Code></Error>`, 404, false, false}, {`denied`, 403, false, false}, {`broken`, 200, false, false},
	} {
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/team:bucket" || r.URL.RawQuery != "object-lock=" || r.Header.Get("Authorization") == "" {
				t.Fatal("wrong signed scoped request")
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
		result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"object-lock"}})
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
		if row["configured"] != tc.configured {
			t.Fatal("wrong configured state")
		}
		if !tc.configured {
			if row["object_lock"] != nil || row["document"] != nil {
				t.Fatal("invented configuration")
			}
			continue
		}
		configuration := row["object_lock"].(s3.BucketObjectLockConfiguration)
		if configuration.DefaultRetention.Mode != "COMPLIANCE" || *configuration.DefaultRetention.Years != "2" || row["document"] != document {
			t.Fatal("lost configuration")
		}
	}
}
