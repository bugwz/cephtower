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

func TestBucketReplicationRead(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	document := `<ReplicationConfiguration><Role/><Rule><ID>rule</ID><Status>Enabled</Status><Destination><Bucket>arn:aws:s3:::target</Bucket><Zone>remote</Zone></Destination><Filter><Prefix>objects/</Prefix></Filter></Rule></ReplicationConfiguration>`
	for _, tc := range []struct {
		body              string
		status            int
		valid, configured bool
		rules             int
	}{
		{document, 200, true, true, 1},
		{"<ReplicationConfiguration><Role/></ReplicationConfiguration>", 200, true, true, 0},
		{`<Error><Code>ReplicationConfigurationNotFoundError</Code></Error>`, 404, true, false, 0},
		{`<Error><Code>NoSuchBucket</Code></Error>`, 404, false, false, 0},
		{`<Error><Code>ReplicationConfigurationNotFoundError</Code></Error>`, 403, false, false, 0},
		{"broken", 200, false, false, 0},
	} {
		for _, tenant := range []string{"", "team"} {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "GET" || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != "replication=" || r.Header.Get("Authorization") == "" {
					t.Fatal("wrong signed scope")
				}
				return &http.Response{StatusCode: tc.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
			result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"replication"}})
			if calls != 1 || (err == nil) != tc.valid {
				t.Fatalf("unexpected result: %v", err)
			}
			if !tc.valid {
				continue
			}
			row := result.(map[string]any)
			if row["configured"] != tc.configured {
				t.Fatal("incorrect presence")
			}
			if !tc.configured {
				if row["replication"] != nil || row["document"] != nil {
					t.Fatal("invented configuration")
				}
				continue
			}
			data := row["replication"].(s3.BucketReplicationConfiguration)
			if data.Rules == nil || len(data.Rules) != tc.rules || row["document"] != tc.body || row["bucket_id"] != id {
				t.Fatal("native data lost")
			}
		}
	}
}
