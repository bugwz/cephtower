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
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestBucketNotificationDeletion(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	wrap := func(s string) string { return "<NotificationConfiguration>" + s + "</NotificationConfiguration>" }
	const target = `<TopicConfiguration><Id> a&amp;+%/中 </Id><Topic>arn:aws:sns:east::topic</Topic><Event>s3:ObjectCreated:*</Event></TopicConfiguration>`
	const other = `<TopicConfiguration><Id>other</Id><Topic>arn:aws:sns:east::other</Topic><Event>future</Event><Filter><S3Tags><FilterRule><Name>key</Name><Value>value</Value></FilterRule></S3Tags></Filter></TopicConfiguration>`
	const selected = " a&+%/中 "
	for _, tenant := range []string{"", "team"} {
		for _, mode := range []string{"single", "all"} {
			for _, scenario := range []string{"success", "last_rule", "invalid_mode", "empty_single", "all_with_id", "missing_snapshot", "changed_snapshot", "pre_denied", "pre_missing", "pre_empty", "pre_broken", "duplicate_id", "unknown_id", "delete_failure", "post_denied", "post_missing", "post_broken", "target_remains", "other_changed"} {
				t.Run(tenant+"/"+mode+"/"+scenario, func(t *testing.T) {
					before := wrap(target + other)
					id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
					chosen := selected
					if mode == "all" {
						chosen = ""
					}
					parameters := map[string]any{"mode": mode, "notification_id": chosen, "expected_document": before}
					switch scenario {
					case "last_rule":
						before = wrap(target)
						parameters["expected_document"] = before
					case "invalid_mode":
						parameters["mode"] = "unknown"
					case "empty_single":
						parameters["mode"], parameters["notification_id"] = "single", ""
					case "all_with_id":
						parameters["mode"], parameters["notification_id"] = "all", selected
					case "missing_snapshot":
						delete(parameters, "expected_document")
					case "changed_snapshot":
						parameters["expected_document"] = wrap(other)
					case "pre_empty":
						before = wrap("")
						parameters["expected_document"] = before
					case "pre_broken":
						before = "broken"
						parameters["expected_document"] = before
					case "duplicate_id":
						before = wrap(target + target + other)
						parameters["expected_document"] = before
					case "unknown_id":
						parameters["mode"], parameters["notification_id"] = "single", "absent"
					}
					calls := 0
					service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
						calls++
						method, query, body, status := "GET", "notification=", before, 200
						if calls == 1 {
							if scenario == "pre_denied" {
								status, body = 403, "secret-native-error"
							}
							if scenario == "pre_missing" {
								status, body = 404, `<Error><Code>NoSuchBucket</Code></Error>`
							}
						} else if calls == 2 {
							method, query, body, status = "DELETE", url.Values{"notification": {chosen}}.Encode(), "", 204
							if scenario == "delete_failure" {
								status, body = 500, "secret-native-error"
							}
						} else {
							body = wrap("")
							if mode == "single" && scenario != "last_rule" {
								body = wrap(other)
							}
							switch scenario {
							case "post_denied":
								status, body = 403, "secret-native-error"
							case "post_missing":
								status, body = 404, `<Error><Code>NoSuchBucket</Code></Error>`
							case "post_broken":
								body = "broken"
							case "target_remains":
								body = before
							case "other_changed":
								body = wrap(strings.Replace(other, "value", "changed", 1))
							}
						}
						if calls > 3 || r.Method != method || r.URL.Path != "/"+tenant+":bucket" || r.URL.RawQuery != strings.ReplaceAll(query, "+", "%20") || !strings.Contains(r.Header.Get("Authorization"), "/s3/aws4_request") {
							t.Fatalf("unexpected native request %d: %s %s", calls, r.Method, r.URL)
						}
						return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.notification_delete", ResourceKey: "rgw/bucket/" + id, Parameters: parameters})
					wantCalls, code := 1, "pre_check_failed"
					switch scenario {
					case "success", "last_rule":
						wantCalls, code = 3, ""
					case "duplicate_id":
						if mode == "all" {
							wantCalls, code = 3, ""
						}
					case "invalid_mode", "empty_single", "all_with_id", "missing_snapshot":
						wantCalls, code = 0, "invalid_request"
					case "delete_failure":
						wantCalls, code = 2, "s3_failed"
					case "post_denied", "post_missing", "post_broken", "target_remains", "other_changed":
						wantCalls, code = 3, "post_check_failed"
					}
					if calls != wantCalls || (err == nil) != (code == "") {
						t.Fatalf("calls %d want %d: %v", calls, wantCalls, err)
					}
					if code != "" {
						var remote *cephdomain.ActionError
						if !errors.As(err, &remote) || remote.Code != code || remote.Retryable || strings.Contains(err.Error(), "secret-native-error") {
							t.Fatalf("unsafe error: %v", err)
						}
					}
				})
			}
		}
	}
}
