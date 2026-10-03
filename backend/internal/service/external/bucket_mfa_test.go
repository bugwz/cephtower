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

func TestBucketMFAStages(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, tenant := range []string{"", "team"} {
		for _, scenario := range []string{"enable", "disable", "never-versioned", "versioning-change", "unchanged", "snapshot", "incomplete", "pre-failed", "write-failed", "post-failed", "wrong-mfa", "wrong-status", "read"} {
			t.Run(tenant+scenario, func(t *testing.T) {
				before := `<VersioningConfiguration><Status>Enabled</Status><MfaDelete>Disabled</MfaDelete></VersioningConfiguration>`
				status, mfa := "Enabled", "Enabled"
				switch scenario {
				case "disable":
					before = strings.Replace(before, "<MfaDelete>Disabled", "<MfaDelete>Enabled", 1)
					mfa = "Disabled"
				case "never-versioned":
					before = "<VersioningConfiguration/>"
				case "versioning-change":
					before = strings.Replace(before, "<MfaDelete>Disabled", "<MfaDelete>Enabled", 1)
					status = "Suspended"
				case "unchanged":
					mfa = "Disabled"
				case "incomplete":
					before = "<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>"
				}
				parameters := map[string]any{"status": status, "mfa_delete": mfa, "mfa_serial_secret": "device:serial", "mfa_token": "001234", "expected_document": before}
				if scenario == "snapshot" {
					parameters["expected_document"] = "changed"
				}
				calls := 0
				service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					method, body, code := "GET", before, 200
					if calls == 1 && scenario == "pre-failed" {
						code, body = 403, "device:serial 001234"
					}
					if calls == 2 {
						method, body = "PUT", ""
						if r.Header.Get("X-Amz-Mfa") != "device:serial 001234" {
							t.Fatal("MFA missing")
						}
						if scenario == "write-failed" {
							code, body = 403, "device:serial 001234"
						}
					} else if r.Header.Get("X-Amz-Mfa") != "" {
						t.Fatal("MFA reused for read")
					}
					if calls == 3 {
						body = "<VersioningConfiguration><Status>" + status + "</Status><MfaDelete>" + mfa + "</MfaDelete></VersioningConfiguration>"
						if scenario == "post-failed" {
							code, body = 404, "device:serial 001234"
						}
						if scenario == "wrong-mfa" {
							body = strings.Replace(body, "<MfaDelete>Enabled", "<MfaDelete>Disabled", 1)
						}
						if scenario == "wrong-status" {
							body = strings.Replace(body, "<Status>Enabled", "<Status>Suspended", 1)
						}
					}
					if calls > 3 || r.Method != method || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != "versioning=" {
						t.Fatal("wrong native request or retry")
					}
					return &http.Response{StatusCode: code, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
				if scenario == "read" {
					result, err := service.Read(ctx, cluster.ID, "rgw_bucket_policy", id, url.Values{"kind": {"versioning"}})
					if err != nil || calls != 1 {
						t.Fatal(err)
					}
					configuration := result.(map[string]any)["versioning"].(s3.BucketVersioningConfiguration)
					if configuration.MFADelete == nil || *configuration.MFADelete != "Disabled" {
						t.Fatal("MFA projection lost")
					}
					return
				}
				_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.mfa", ResourceKey: "rgw/bucket/" + id, Parameters: parameters})
				wantCalls, success := 3, scenario == "enable" || scenario == "disable" || scenario == "never-versioned" || scenario == "versioning-change"
				if scenario == "unchanged" || scenario == "snapshot" || scenario == "incomplete" || scenario == "pre-failed" {
					wantCalls = 1
				}
				if scenario == "write-failed" {
					wantCalls = 2
				}
				if calls != wantCalls || (err == nil) != success {
					t.Fatalf("calls %d, want %d: %v", calls, wantCalls, err)
				}
				if err != nil {
					var action *cephdomain.ActionError
					if !errors.As(err, &action) || action.Retryable || strings.Contains(err.Error(), "001234") || strings.Contains(err.Error(), "device:serial") {
						t.Fatalf("unsafe error: %v", err)
					}
				}
			})
		}
	}
}
