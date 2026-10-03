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

func TestTopicAttributeVerification(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		for _, tc := range []struct{ attribute, old, value, wire string }{
			{"OpaqueData", "old", "中文 &+<>\n", "中文 &+<>\n"}, {"OpaqueData", "old", "", ""},
			{"persistent", "true", "false", "false"}, {"persistent", "false", "true", "true"},
			{"time_to_live", "None", "0", "0"}, {"time_to_live", "0", "None", "-1"},
			{"max_retries", "10", "2147483647", "2147483647"}, {"max_retries", "10", "None", "-1"},
			{"retry_sleep_duration", "0", "1", "1"}, {"retry_sleep_duration", "1", "None", "-1"},
		} {
			for _, mode := range []string{"success", "stale", "write error", "post error", "other destination changed", "policy changed", "value unchanged", "missing field", "null field", "identity"} {
				t.Run(scope+"/"+tc.attribute+"/"+tc.value+"/"+mode, func(t *testing.T) {
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
							t.Fatal("wrong ARN")
						}
						status := 200
						body := ""
						if calls == 2 {
							if r.Form.Get("Action") != "SetTopicAttributes" || r.Form.Get("AttributeName") != tc.attribute || r.Form.Get("AttributeValue") != tc.wire || !r.Form.Has("AttributeValue") {
								t.Fatalf("wrong write: %v", r.Form)
							}
							body = "<SetTopicAttributesResponse/>"
							if mode == "write error" {
								status = 500
								body = "secret-password"
							}
						} else {
							if r.Form.Get("Action") != "GetTopicAttributes" {
								t.Fatal("wrong read")
							}
							dest := map[string]any{"EndpointAddress": "https://user:secret-password@host/path", "EndpointArgs": "password=hidden", "EndpointTopic": "events", "HasStoredSecret": true, "Persistent": true, "TimeToLive": "None", "MaxRetries": "0", "RetrySleepDuration": "10", "FutureCounter": json.Number("9007199254740993")}
							attrs := map[string]string{"User": "owner", "Name": "events", "TopicArn": arn, "OpaqueData": "old", "Policy": "{}"}
							value := tc.old
							if calls == 3 && mode != "value unchanged" {
								value = tc.value
							}
							if mode == "stale" {
								value = "stale"
							}
							if tc.attribute == "OpaqueData" {
								attrs["OpaqueData"] = value
							} else if tc.attribute == "persistent" {
								if value == "stale" {
									dest["Persistent"] = "unknown"
								} else {
									dest["Persistent"] = value == "true"
								}
							} else {
								dest[topicDestinationField(tc.attribute)] = value
							}
							if mode == "identity" {
								attrs["TopicArn"] = "other"
							}
							if mode == "missing field" {
								delete(dest, "HasStoredSecret")
							}
							if mode == "null field" {
								dest["HasStoredSecret"] = nil
							}
							if calls == 3 && mode == "other destination changed" {
								dest["EndpointArgs"] = "password=changed"
							}
							if calls == 3 && mode == "policy changed" {
								attrs["Policy"] = "changed"
							}
							encoded, _ := json.Marshal(dest)
							attrs["EndPoint"] = string(encoded)
							var b strings.Builder
							b.WriteString("<GetTopicAttributesResponse><GetTopicAttributesResult><Attributes>")
							for k, v := range attrs {
								b.WriteString("<entry><key>" + k + "</key><value>")
								_ = xml.EscapeText(&b, []byte(v))
								b.WriteString("</value></entry>")
							}
							b.WriteString("</Attributes></GetTopicAttributesResult></GetTopicAttributesResponse>")
							body = b.String()
							if calls == 3 && mode == "post error" {
								status = 403
								body = "secret-password"
							}
						}
						return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_topic.attribute", ResourceKey: "rgw/topic/" + id, Parameters: map[string]any{"topic_id": id, "topic_arn": arn, "attribute": tc.attribute, "expected_value": tc.old, "value": tc.value}})
					wantCalls := 3
					code := "post_check_failed"
					switch mode {
					case "success":
						code = ""
					case "stale", "missing field", "null field", "identity":
						wantCalls = 1
						code = "pre_check_failed"
					case "write error":
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
						if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || strings.Contains(err.Error(), "secret-password") {
							t.Fatalf("unsafe error %v", err)
						}
					}
				})
			}
		}
	}
}

func TestTopicAttributeRejectsInvalidBeforeHTTP(t *testing.T) {
	service, _, cluster := externalTestService(t)
	for _, tc := range []struct{ attribute, value string }{
		{"Policy", "{}"}, {"push-endpoint", "https://other"}, {"password", "secret"}, {"persistent", "1"},
		{"time_to_live", "2147483648"}, {"time_to_live", "-1"}, {"max_retries", "01"}, {"retry_sleep_duration", "1.5"}, {"OpaqueData", "old"},
	} {
		t.Run(tc.attribute+tc.value, func(t *testing.T) {
			id := base64.RawURLEncoding.EncodeToString([]byte(":events"))
			_, err := service.Execute(context.Background(), Request{ClusterID: cluster.ID, Action: "rgw_topic.attribute", ResourceKey: "rgw/topic/" + id, Parameters: map[string]any{"topic_id": id, "topic_arn": "arn:aws:sns:default::events", "attribute": tc.attribute, "expected_value": "old", "value": tc.value}})
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Code != "invalid_request" {
				t.Fatalf("invalid input reached endpoint: %v", err)
			}
		})
	}
}
