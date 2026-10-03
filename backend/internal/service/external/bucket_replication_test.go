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

func TestBucketReplicationDeletionVerifiesEmptyRules(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	populated := `<ReplicationConfiguration><Role/><Rule><ID>r</ID><Status>Enabled</Status><Destination><Bucket>target</Bucket></Destination></Rule></ReplicationConfiguration>`
	empty := "<ReplicationConfiguration><Role/></ReplicationConfiguration>"
	missing := "<Error><Code>ReplicationConfigurationNotFoundError</Code></Error>"
	for _, tc := range []struct {
		name, before, after                        string
		preStatus, deleteStatus, postStatus, calls int
		code                                       string
	}{
		{"empty after", populated, empty, 200, 204, 200, 3, ""},
		{"missing after", populated, missing, 200, 204, 404, 3, ""},
		{"already empty", empty, empty, 200, 204, 200, 1, "pre_check_failed"},
		{"missing before", missing, empty, 404, 204, 200, 1, "pre_check_failed"},
		{"pre denied", populated, empty, 403, 204, 200, 1, "pre_check_failed"},
		{"pre malformed", "broken", empty, 200, 204, 200, 1, "pre_check_failed"},
		{"delete failed", populated, empty, 200, 503, 200, 2, "s3_failed"},
		{"delete uncertain", populated, empty, 200, 0, 200, 2, "s3_failed"},
		{"still present", populated, populated, 200, 204, 200, 3, "post_check_failed"},
		{"disabled still present", populated, strings.Replace(populated, "Enabled", "Disabled", 1), 200, 204, 200, 3, "post_check_failed"},
		{"post malformed", populated, "broken", 200, 204, 200, 3, "post_check_failed"},
		{"post denied", populated, missing, 200, 204, 403, 3, "post_check_failed"},
		{"wrong missing code", populated, "<Error><Code>NoSuchBucket</Code></Error>", 200, 204, 404, 3, "post_check_failed"},
	} {
		for _, tenant := range []string{"", "team"} {
			t.Run(tc.name+"/"+tenant, func(t *testing.T) {
				calls := 0
				service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					method, body, status := "GET", tc.before, tc.preStatus
					if calls == 2 {
						method, body, status = "DELETE", "", tc.deleteStatus
					}
					if calls == 3 {
						body, status = tc.after, tc.postStatus
					}
					if calls > 3 || r.Method != method || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != "replication=" || r.Header.Get("Authorization") == "" {
						t.Fatal("wrong signed scope or sequence")
					}
					if status == 0 {
						return nil, errors.New("connection lost")
					}
					return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
				_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.delete", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": "replication"}})
				if calls != tc.calls {
					t.Fatalf("calls=%d want=%d", calls, tc.calls)
				}
				if tc.code == "" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				var actionError *cephdomain.ActionError
				if !errors.As(err, &actionError) || actionError.Code != tc.code || actionError.Retryable {
					t.Fatalf("wrong failure: %v", err)
				}
			})
		}
	}
}
