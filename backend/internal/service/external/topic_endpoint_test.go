package external

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestTopicEndpointVerifiedReplacement(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		for _, tc := range []struct {
			old, value string
			stored     bool
		}{
			{"https://old/path", "https://new/path?q=private&x=1", false},
			{"https://old/path", "amqps://user:secret-password@broker/vhost", false},
			{"amqp://user:old-secret@old/vhost", "kafka://new:9092", true},
			{"https://user:old-secret@old/path?q=hidden", "", true},
			{"", "https://new/path", false},
			{"https://old/path", "http://user:secret-password@new/path", true},
		} {
			for _, mode := range []string{"success", "stale", "write failure", "post failure", "wrong endpoint", "args changed", "policy changed", "stored marker changed", "identity", "missing field", "unchanged", "stale marker", "stale redaction"} {
				t.Run(scope+"/"+tc.value+"/"+mode, func(t *testing.T) {
					arn := "arn:aws:sns:default:" + scope + ":events"
					id := base64.RawURLEncoding.EncodeToString([]byte(scope + ":events"))
					calls := 0
					expected, redacted, _ := s3.RedactTopicEndpoint(tc.old)
					service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
						calls++
						if r.Method != "POST" || r.URL.RawQuery != "" || !strings.Contains(r.Header.Get("Authorization"), "/sns/aws4_request") {
							t.Fatal("wrong SNS request")
						}
						if err := r.ParseForm(); err != nil {
							t.Fatal(err)
						}
						if r.Form.Get("TopicArn") != arn {
							t.Fatal("wrong scoped target")
						}
						status := 200
						body := ""
						if calls == 2 {
							if r.Form.Get("Action") != "SetTopicAttributes" || r.Form.Get("AttributeName") != "push-endpoint" || r.Form.Get("AttributeValue") != tc.value || !r.Form.Has("AttributeValue") {
								t.Fatal("wrong endpoint write")
							}
							body = "<SetTopicAttributesResponse/>"
							if mode == "write failure" {
								status = 400
								body = "secret-password"
							}
						} else {
							if r.Form.Get("Action") != "GetTopicAttributes" {
								t.Fatal("wrong read")
							}
							value := tc.old
							stored := tc.stored
							if calls == 3 {
								value = tc.value
								secret, _ := s3.ValidateTopicEndpoint(tc.value)
								stored = stored || secret
							}
							if mode == "stale" {
								value = "https://other/path"
							}
							if calls == 3 && mode == "wrong endpoint" {
								value = tc.old
							}
							if calls == 3 && mode == "stored marker changed" {
								stored = !stored
							}
							dest := map[string]any{"EndpointAddress": value, "EndpointArgs": "password=retained-secret&amqp-exchange=exchange", "EndpointTopic": "events", "Persistent": true, "HasStoredSecret": stored, "TimeToLive": "None", "MaxRetries": "0", "RetrySleepDuration": "10", "FutureCounter": json.Number("9007199254740993")}
							if calls == 3 && mode == "args changed" {
								dest["EndpointArgs"] = ""
							}
							if mode == "missing field" {
								delete(dest, "Persistent")
							}
							encoded, _ := json.Marshal(dest)
							attrs := map[string]string{"User": "owner", "Name": "events", "TopicArn": arn, "OpaqueData": "opaque", "Policy": "{}", "EndPoint": string(encoded)}
							if calls == 3 && mode == "policy changed" {
								attrs["Policy"] = "changed"
							}
							if mode == "identity" {
								attrs["TopicArn"] = "other"
							}
							var b strings.Builder
							b.WriteString("<GetTopicAttributesResponse><GetTopicAttributesResult><Attributes>")
							for k, v := range attrs {
								b.WriteString("<entry><key>" + k + "</key><value>")
								_ = xml.EscapeText(&b, []byte(v))
								b.WriteString("</value></entry>")
							}
							b.WriteString("</Attributes></GetTopicAttributesResult></GetTopicAttributesResponse>")
							body = b.String()
							if calls == 3 && mode == "post failure" {
								status = 403
								body = "secret-password"
							}
						}
						return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					parameters := map[string]any{"topic_id": id, "topic_arn": arn, "expected_endpoint": expected, "expected_redacted": redacted, "expected_stored_secret": tc.stored, "endpoint_secret": tc.value}
					if mode == "unchanged" {
						parameters["endpoint_secret"] = tc.old
					}
					if mode == "stale marker" {
						parameters["expected_stored_secret"] = !tc.stored
					}
					if mode == "stale redaction" {
						parameters["expected_redacted"] = !redacted
					}
					_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_topic.endpoint", ResourceKey: "rgw/topic/" + id, Parameters: parameters})
					wantCalls := 3
					code := "post_check_failed"
					switch mode {
					case "success":
						code = ""
					case "stale", "identity", "missing field", "unchanged", "stale marker", "stale redaction":
						wantCalls = 1
						code = "pre_check_failed"
					case "write failure":
						wantCalls = 2
						code = "sns_failed"
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
						if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || strings.Contains(err.Error(), "secret-password") || strings.Contains(err.Error(), "retained-secret") {
							t.Fatalf("unsafe failure %v", err)
						}
					}
				})
			}
		}
	}
}
