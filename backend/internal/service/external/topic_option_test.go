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
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestTopicOptionVerifiedWrites(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		for option, value := range map[string]string{"verify-ssl": "false", "use-ssl": "true", "cloudevents": "true", "ca-location": "/etc/ca.pem", "amqp-exchange": "events", "amqp-ack-level": "routable", "kafka-ack-level": "none", "mechanism": "SCRAM-SHA-256"} {
			for _, mode := range []string{"success", "unsafe match", "write failure", "post failure", "other args changed", "policy changed"} {
				t.Run(scope+option+mode, func(t *testing.T) {
					arn := "arn:aws:sns:default:" + scope + ":events"
					id := base64.RawURLEncoding.EncodeToString([]byte(scope + ":events"))
					calls := 0
					service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
						calls++
						if r.Method != "POST" || r.URL.RawQuery != "" || !strings.Contains(r.Header.Get("Authorization"), "/sns/aws4_request") {
							t.Fatal("wrong SNS request")
						}
						if err := r.ParseForm(); err != nil {
							t.Fatal(err)
						}
						if r.Form.Get("TopicArn") != arn {
							t.Fatal("wrong scope")
						}
						status := 200
						body := ""
						if calls == 2 {
							if r.Form.Get("Action") != "SetTopicAttributes" || r.Form.Get("AttributeName") != option || r.Form.Get("AttributeValue") != value {
								t.Fatal("wrong write")
							}
							body = "<SetTopicAttributesResponse/>"
							if mode == "write failure" {
								status = 500
								body = "sensitive-secret"
							}
						} else {
							if r.Form.Get("Action") != "GetTopicAttributes" {
								t.Fatal("wrong read")
							}
							raw := "password=sensitive-secret&unknown=kept"
							if mode == "unsafe match" {
								raw = "password=" + option + "-sensitive-secret"
							}
							if calls == 3 {
								raw += "&" + option + "=" + value
							}
							if calls == 3 && mode == "other args changed" {
								raw = "password=changed&" + option + "=" + value
							}
							dest := map[string]any{"EndpointAddress": "https://host/path", "EndpointArgs": raw, "EndpointTopic": "events", "HasStoredSecret": true, "Persistent": false, "TimeToLive": "None", "MaxRetries": "0", "RetrySleepDuration": "10"}
							encoded, _ := json.Marshal(dest)
							attrs := map[string]string{"User": "owner", "Name": "events", "TopicArn": arn, "OpaqueData": "opaque", "Policy": "{}", "EndPoint": string(encoded)}
							if calls == 3 && mode == "policy changed" {
								attrs["Policy"] = "changed"
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
								body = "sensitive-secret"
							}
						}
						return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_topic.option", ResourceKey: "rgw/topic/" + id, Parameters: map[string]any{"topic_id": id, "topic_arn": arn, "option": option, "value": value, "expected_status": "unset", "expected_value": ""}})
					code := "post_check_failed"
					wantCalls := 3
					switch mode {
					case "success":
						code = ""
					case "unsafe match":
						code = "pre_check_failed"
						wantCalls = 1
					case "write failure":
						code = "sns_failed"
						wantCalls = 2
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
						if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || strings.Contains(err.Error(), "sensitive-secret") {
							t.Fatalf("unsafe failure %v", err)
						}
					}
				})
			}
		}
	}
}
