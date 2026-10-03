package external

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"encoding/base64"
	"errors"
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

func TestObjectLockWriteSequence(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	empty := `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled></ObjectLockConfiguration>`
	sent := `<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>GOVERNANCE</Mode><Days>030</Days></DefaultRetention></Rule></ObjectLockConfiguration>`
	for _, tc := range []struct {
		name                               string
		initialStatus                      int
		initial, version, document, actual string
		putStatus                          int
		code                               string
		calls                              int
	}{
		{"update", 200, empty, "", sent, strings.Replace(sent, "030", "30", 1), 200, "", 3},
		{"clear default", 200, sent, "", empty, empty, 200, "", 3},
		{"enable", 404, `<Error><Code>ObjectLockConfigurationNotFoundError</Code></Error>`, "Enabled", sent, strings.Replace(sent, "030", "30", 1), 200, "", 4},
		{"suspended", 404, `<Error><Code>ObjectLockConfigurationNotFoundError</Code></Error>`, "Suspended", sent, "", 200, "pre_check_failed", 2},
		{"wrong bucket", 404, `<Error><Code>NoSuchBucket</Code></Error>`, "", sent, "", 200, "pre_check_failed", 1},
		{"preflight denied", 403, "denied", "", sent, "", 200, "pre_check_failed", 1},
		{"write failed", 200, empty, "", sent, "", 503, "s3_failed", 2},
		{"changed period", 200, empty, "", sent, strings.Replace(sent, "030", "31", 1), 200, "post_check_failed", 3},
		{"bad readback", 200, empty, "", sent, "broken", 200, "post_check_failed", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != "/team:bucket" {
					t.Fatal("wrong tenant")
				}
				status, body := 200, tc.actual
				if calls == 1 {
					status, body = tc.initialStatus, tc.initial
					if r.Method != "GET" {
						t.Fatal("missing preflight")
					}
				} else if tc.version != "" && calls == 2 {
					if r.Method != "GET" || r.URL.RawQuery != "versioning=" {
						t.Fatal("missing versioning preflight")
					}
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("<VersioningConfiguration><Status>" + tc.version + "</Status></VersioningConfiguration>"))}, nil
				} else if r.Method == "PUT" {
					data, _ := io.ReadAll(r.Body)
					if string(data) != tc.document {
						t.Fatal("changed write document")
					}
					status, body = tc.putStatus, ""
				} else if r.Method != "GET" {
					t.Fatal("unexpected method")
				}
				if r.URL.RawQuery != "object-lock=" {
					t.Fatal("wrong configuration target")
				}
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			id := base64.RawURLEncoding.EncodeToString([]byte("team\x00bucket"))
			_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.update", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": "object-lock", "document": tc.document}})
			if calls != tc.calls {
				t.Fatalf("calls=%d expected=%d err=%v", calls, tc.calls, err)
			}
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Code != tc.code || actionError.Retryable {
				t.Fatalf("unsafe error %v", err)
			}
		})
	}
}
