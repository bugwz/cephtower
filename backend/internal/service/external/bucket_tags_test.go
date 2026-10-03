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

func TestBucketTagsReadWriteAndVerification(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
	document := `<Tagging><TagSet><Tag><Key>a</Key><Value>2</Value></Tag><Tag><Key>a</Key><Value>1</Value></Tag></TagSet></Tagging>`
	reordered := `<Tagging><TagSet><Tag><Key>a</Key><Value>1</Value></Tag><Tag><Key>a</Key><Value>2</Value></Tag></TagSet></Tagging>`
	empty := `<Tagging><TagSet/></Tagging>`
	for _, tc := range []struct {
		body              string
		status            int
		valid, configured bool
		count             int
	}{
		{document, 200, true, true, 2}, {empty, 200, true, true, 0}, {`<Error><Code>NoSuchTagSet</Code></Error>`, 404, true, false, 0},
		{`<Error><Code>NoSuchBucket</Code></Error>`, 404, false, false, 0}, {`<Tagging/>`, 200, false, false, 0},
	} {
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/team:bucket" || r.URL.RawQuery != "tagging=" {
				t.Fatalf("wrong read %s", r.URL)
			}
			return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": {"application/xml"}}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})
		result, err := service.readBucketPolicy(ctx, cluster.ID, id, url.Values{"kind": {"tagging"}})
		if !tc.valid {
			if err == nil {
				t.Fatal("invalid tags accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		row := result.(map[string]any)
		if row["configured"] != tc.configured || len(row["tags"].([]s3.BucketTag)) != tc.count {
			t.Fatalf("wrong tags: %+v", row)
		}
	}
	for _, tc := range []struct {
		sent, returned       string
		putStatus, getStatus int
		code                 string
	}{
		{document, reordered, 200, 200, ""}, {empty, empty, 200, 200, ""}, {document, empty, 200, 200, "post_check_failed"},
		{document, `<Tagging/>`, 200, 200, "post_check_failed"}, {document, `<Error><Code>NoSuchTagSet</Code></Error>`, 200, 404, "post_check_failed"},
		{document, document, 200, 403, "post_check_failed"}, {document, document, 503, 200, "s3_failed"}, {`<Tagging/>`, document, 200, 200, "invalid_request"},
	} {
		calls := 0
		service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Path != "/team:bucket" || r.URL.RawQuery != "tagging=" {
				t.Fatalf("wrong target %s", r.URL)
			}
			status, body := tc.getStatus, tc.returned
			if calls == 1 {
				data, _ := io.ReadAll(r.Body)
				if r.Method != "PUT" || string(data) != tc.sent || r.Header.Get("Content-Type") != "application/xml" || r.Header.Get("Content-MD5") == "" {
					t.Fatal("wrong native tagging write")
				}
				status, body = tc.putStatus, ""
			} else if calls != 2 || r.Method != "GET" {
				t.Fatal("unexpected verification request")
			}
			return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
		})
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": "tagging", "document": tc.sent}})
		wantCalls := 2
		if tc.code == "invalid_request" {
			wantCalls = 0
		}
		if tc.code == "s3_failed" {
			wantCalls = 1
		}
		if calls != wantCalls {
			t.Fatalf("calls=%d want=%d", calls, wantCalls)
		}
		if tc.code == "" {
			if err != nil {
				t.Fatal(err)
			}
			continue
		}
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != tc.code || failure.Retryable {
			t.Fatalf("unsafe failure: %v", err)
		}
	}
}
