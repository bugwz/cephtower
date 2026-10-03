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

func TestBucketACLWriteSequence(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	owner := `<Grant><Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser"><ID>owner</ID></Grantee><Permission>FULL_CONTROL</Permission></Grant>`
	wrap := func(grants string) string {
		return `<AccessControlPolicy><Owner><ID>owner</ID></Owner><AccessControlList>` + grants + `</AccessControlList></AccessControlPolicy>`
	}
	group := func(uri, permissions string) string {
		return `<Grant><Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="Group"><URI>http://acs.amazonaws.com/groups/global/` + uri + `</URI></Grantee>` + permissions + `</Grant>`
	}
	read := "<Permission>READ</Permission>"
	private := wrap(owner)
	public := wrap(owner + group("AllUsers", read))
	for _, tc := range []struct {
		name, canned, before, after             string
		preStatus, putStatus, postStatus, calls int
		code                                    string
	}{
		{"private", "private", public, private, 200, 200, 200, 3, ""},
		{"public read", "public-read", private, public, 200, 200, 200, 3, ""},
		{"public write", "public-read-write", private, wrap(owner + group("AllUsers", read+"<Permission>WRITE</Permission>")), 200, 200, 200, 3, ""},
		{"authenticated", "authenticated-read", private, wrap(owner + group("AuthenticatedUsers", read)), 200, 200, 200, 3, ""},
		{"bad preset", "other", private, private, 200, 200, 200, 0, "invalid_request"},
		{"pre denied", "private", private, private, 403, 200, 200, 1, "pre_check_failed"},
		{"pre malformed", "private", "broken", private, 200, 200, 200, 1, "pre_check_failed"},
		{"write rejected", "private", private, private, 200, 403, 200, 2, "s3_failed"},
		{"write uncertain", "private", private, private, 200, 0, 200, 2, "s3_failed"},
		{"post denied", "private", private, private, 200, 200, 403, 3, "post_check_failed"},
		{"post malformed", "private", private, "broken", 200, 200, 200, 3, "post_check_failed"},
		{"extra grant", "private", private, public, 200, 200, 200, 3, "post_check_failed"},
		{"owner changed", "private", private, strings.ReplaceAll(private, ">owner<", ">other<"), 200, 200, 200, 3, "post_check_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != "/team:bucket" || r.URL.RawQuery != "acl=" || r.Header.Get("Authorization") == "" {
					t.Fatal("scope or signature missing")
				}
				body, status := tc.before, tc.preStatus
				if calls == 2 {
					if r.Method != "PUT" || r.Header.Get("X-Amz-Acl") != tc.canned || !strings.Contains(r.Header.Get("Authorization"), "SignedHeaders=host;x-amz-acl;") {
						t.Fatal("ACL header not signed")
					}
					sent, _ := io.ReadAll(r.Body)
					if len(sent) != 0 {
						t.Fatal("canned ACL must not send XML body")
					}
					status = tc.putStatus
					if status == 0 {
						return nil, errors.New("connection lost")
					}
				} else {
					if r.Method != "GET" || r.Header.Get("X-Amz-Acl") != "" {
						t.Fatal("invalid read")
					}
					if calls == 3 {
						body, status = tc.after, tc.postStatus
					}
				}
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
			_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.acl", ResourceKey: "rgw/bucket/" + id, Parameters: map[string]any{"acl": tc.canned}})
			if calls != tc.calls {
				t.Fatalf("calls: %d, want %d", calls, tc.calls)
			}
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Code != tc.code || failure.Retryable {
				t.Fatalf("unexpected error: %#v", err)
			}
		})
	}
}
