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

func TestBucketNotificationRead(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		body          string
		status, count int
		valid         bool
	}{
		{`<NotificationConfiguration><TopicConfiguration><Id>rule</Id><Topic>arn:aws:sns:east::topic</Topic><Event>s3:ObjectCreated:*</Event></TopicConfiguration></NotificationConfiguration>`, 200, 1, true},
		{`<NotificationConfiguration/>`, 200, 0, true},
		{`<Error><Code>NoSuchBucket</Code></Error>`, 404, 0, false},
		{`<Error><Code>NoSuchKey</Code></Error>`, 404, 0, false},
		{`<Error><Code>AccessDenied</Code></Error>`, 403, 0, false},
		{`<NotificationConfiguration/>`, 503, 0, false},
		{`broken`, 200, 0, false},
	} {
		for _, tenant := range []string{"", "team"} {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "GET" || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != "notification=" || !strings.Contains(r.Header.Get("Authorization"), "/s3/aws4_request") {
					t.Fatal("wrong signed native notification target")
				}
				return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": {"application/xml"}}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
			result, err := service.Read(ctx, cluster.ID, "rgw_bucket_policy", id, url.Values{"kind": {"notification"}})
			if calls != 1 || (err == nil) != tc.valid {
				t.Fatalf("unexpected result: %v", err)
			}
			if tc.valid {
				row := result.(map[string]any)
				rules := row["notifications"].([]s3.BucketNotification)
				if row["configured"] != true || row["document"] != tc.body || row["bucket_id"] != id || row["content_type"] != "application/xml" || rules == nil || len(rules) != tc.count {
					t.Fatalf("native data lost: %+v", row)
				}
			}
		}
	}
}
