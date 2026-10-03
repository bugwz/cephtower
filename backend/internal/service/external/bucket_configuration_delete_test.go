package external

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestBucketConfigurationDeletionVerifiesNativeAbsence(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	if !Supports("rgw_bucket_policy.delete") {
		t.Fatal("dispatcher cannot route deletion")
	}
	for kind, code := range map[string]string{"policy": "NoSuchBucketPolicy", "cors": "NoSuchCORSConfiguration", "lifecycle": "NoSuchLifecycleConfiguration", "encryption": "ServerSideEncryptionConfigurationNotFoundError", "tagging": "NoSuchTagSet"} {
		for _, tenant := range []string{"", "team"} {
			for _, scenario := range []string{"success", "pre-denied", "pre-missing", "delete-failed", "delete-transport", "post-present", "post-denied", "post-wrong-code", "post-invalid", "post-transport"} {
				t.Run(kind+"/"+tenant+"/"+scenario, func(t *testing.T) {
					calls := 0
					service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
						calls++
						method := "GET"
						if calls == 2 {
							method = "DELETE"
						}
						if calls > 3 || r.Method != method || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != kind+"=" || r.Header.Get("Authorization") == "" {
							t.Fatalf("unsafe request: %s %s", r.Method, r.URL)
						}
						body, _ := io.ReadAll(r.Body)
						if len(body) != 0 {
							t.Fatal("delete sent a document")
						}
						status, response := 200, "native configuration"
						if kind == "tagging" {
							response = "<Tagging><TagSet/></Tagging>"
						}
						if calls == 1 && scenario == "pre-denied" {
							status = 403
						}
						if calls == 1 && scenario == "pre-missing" {
							status, response = 404, "<Error><Code>"+code+"</Code></Error>"
						}
						if calls == 2 {
							status, response = 204, ""
							if scenario == "delete-failed" {
								status = 503
							}
							if scenario == "delete-transport" {
								return nil, errors.New("connection lost")
							}
						}
						if calls == 3 {
							status, response = 404, "<Error><Code>"+code+"</Code></Error>"
							switch scenario {
							case "post-present":
								status, response = 200, "still configured"
								if kind == "tagging" {
									response = "<Tagging><TagSet/></Tagging>"
								}
							case "post-denied":
								status = 403
							case "post-wrong-code":
								response = "<Error><Code>NoSuchBucket</Code></Error>"
							case "post-invalid":
								response = "<Error><Code>" + code
							case "post-transport":
								return nil, errors.New("connection lost")
							}
						}
						return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(response))}, nil
					})
					id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
					_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.delete", ResourceKey: "rgw/bucket/" + id + "/policy", Parameters: map[string]any{"kind": kind}})
					wantCalls, wantCode := 3, "post_check_failed"
					if strings.HasPrefix(scenario, "pre-") {
						wantCalls, wantCode = 1, "pre_check_failed"
					}
					if strings.HasPrefix(scenario, "delete-") {
						wantCalls, wantCode = 2, "s3_failed"
					}
					if calls != wantCalls {
						t.Fatalf("calls=%d want=%d", calls, wantCalls)
					}
					if scenario == "success" {
						if err != nil {
							t.Fatal(err)
						}
						return
					}
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Code != wantCode || failure.Retryable {
						t.Fatalf("unsafe result: %v", err)
					}
				})
			}
		}
	}
	service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid kind reached network")
		return nil, nil
	})
	for _, kind := range []any{"", "versioning", "unknown", nil, true} {
		_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket_policy.delete", ResourceKey: "rgw/bucket/AGJ1Y2tldA/policy", Parameters: map[string]any{"kind": kind}})
		if err == nil {
			t.Fatal("accepted invalid kind")
		}
	}
}
