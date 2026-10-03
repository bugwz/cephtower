package external

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTopicPolicyVerifiedSNS(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret", "session_token": "token", "region": "default"}}); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		for _, mode := range []string{"set", "clear", "stale", "wrong identity", "write failure", "post policy", "post endpoint", "duplicate", "missing", "invalid", "unchanged"} {
			t.Run(scope+"/"+mode, func(t *testing.T) {
				arn := "arn:aws:sns:default:" + scope + ":events"
				id := base64.RawURLEncoding.EncodeToString([]byte(scope + ":events"))
				expected := `{"Statement":[]}`
				desired := `{"Statement":[],"Id":"中文&+<>"}`
				if mode == "clear" {
					desired = ""
				}
				if mode == "invalid" {
					desired = "[]"
				}
				if mode == "unchanged" {
					desired = expected
				}
				calls := 0
				service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if r.Method != "POST" || r.URL.Path != "/" || r.URL.RawQuery != "" || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || !strings.Contains(r.Header.Get("Authorization"), "/sns/aws4_request") || r.Header.Get("X-Amz-Security-Token") != "token" {
						t.Fatalf("incorrect signed SNS request: %s", r.URL)
					}
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if r.Form.Get("TopicArn") != arn || r.Form.Get("Version") != "2010-03-31" {
						t.Fatal("wrong target")
					}
					status := 200
					body := ""
					if calls == 2 {
						if r.Form.Get("Action") != "SetTopicAttributes" || r.Form.Get("AttributeName") != "Policy" || r.Form.Get("AttributeValue") != desired || !r.Form.Has("AttributeValue") {
							t.Fatal("wrong mutation")
						}
						body = "<SetTopicAttributesResponse><ResponseMetadata/></SetTopicAttributesResponse>"
						if mode == "write failure" {
							status = 403
							body = "sensitive remote body"
						}
					} else {
						if r.Form.Get("Action") != "GetTopicAttributes" {
							t.Fatal("wrong read")
						}
						policy := expected
						if calls == 3 {
							policy = desired
						}
						if mode == "stale" || mode == "post policy" && calls == 3 {
							policy = "different"
						}
						attrs := map[string]string{"User": "owner", "Name": "events", "EndPoint": `{"push_endpoint":"https://secret@example.test"}`, "TopicArn": arn, "OpaqueData": "opaque", "Policy": policy}
						if mode == "wrong identity" {
							attrs["TopicArn"] = "other"
						}
						if mode == "post endpoint" && calls == 3 {
							attrs["EndPoint"] = "{}"
						}
						if mode == "missing" {
							delete(attrs, "Policy")
						}
						var b strings.Builder
						b.WriteString("<GetTopicAttributesResponse><GetTopicAttributesResult><Attributes>")
						for k, v := range attrs {
							b.WriteString("<entry><key>" + k + "</key><value>")
							_ = xml.EscapeText(&b, []byte(v))
							b.WriteString("</value></entry>")
						}
						if mode == "duplicate" {
							b.WriteString("<entry><key>Name</key><value>events</value></entry>")
						}
						b.WriteString("</Attributes></GetTopicAttributesResult><ResponseMetadata/></GetTopicAttributesResponse>")
						body = b.String()
					}
					return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_topic.policy", ResourceKey: "rgw/topic/" + id, Parameters: map[string]any{"topic_id": id, "topic_arn": arn, "expected_policy": expected, "policy": desired}})
				code := ""
				wantCalls := 3
				switch mode {
				case "stale", "wrong identity", "duplicate", "missing":
					code = "pre_check_failed"
					wantCalls = 1
				case "write failure":
					code = "sns_failed"
					wantCalls = 2
				case "post policy", "post endpoint":
					code = "post_check_failed"
				case "invalid", "unchanged":
					code = "invalid_request"
					wantCalls = 0
				}
				if calls != wantCalls {
					t.Fatalf("calls %d want %d", calls, wantCalls)
				}
				if code == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || strings.Contains(err.Error(), "sensitive") {
						t.Fatalf("unsafe error: %v", err)
					}
				}
			})
		}
	}
}
