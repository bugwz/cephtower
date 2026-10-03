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

type topicOwnerVerifierFunc func(context.Context, uint64, string, string, string) (string, string, error)

func (f topicOwnerVerifierFunc) VerifyTopicOwner(ctx context.Context, id uint64, uid, key, secret string) (string, string, error) {
	return f(ctx, id, uid, key, secret)
}
func TestTopicCreationVerifiedStages(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "sensitive-secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		for _, mode := range []string{"success", "wrong owner", "exists", "denied", "ambiguous absence", "wrong absence code", "create failure", "wrong returned ARN", "post denied", "wrong field", "wrong args", "wrong secret flag", "wrong owner readback"} {
			t.Run(scope+mode, func(t *testing.T) {
				arn := "arn:aws:sns:zone:" + scope + ":events"
				id := base64.RawURLEncoding.EncodeToString([]byte(scope + ":events"))
				calls, owners := 0, 0
				owner := "user"
				if strings.HasPrefix(scope, "RGW") {
					owner = scope
				} else if scope != "" {
					owner = scope + "$user"
				}
				service.topicOwners = topicOwnerVerifierFunc(func(_ context.Context, cid uint64, uid, key, secret string) (string, string, error) {
					owners++
					if cid != cluster.ID || uid != "user" || key != "access" || secret != "sensitive-secret" {
						t.Fatal("wrong verification input")
					}
					if mode == "wrong owner" {
						return "wrong", owner, nil
					}
					return scope, owner, nil
				})
				service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					if r.Method != "POST" || r.URL.RawQuery != "" || !strings.Contains(r.Header.Get("Authorization"), "/sns/aws4_request") {
						t.Fatal("wrong SNS transport")
					}
					if err := r.ParseForm(); err != nil {
						t.Fatal(err)
					}
					status := 200
					body := ""
					if calls == 1 {
						if r.Form.Get("Action") != "GetTopicAttributes" || r.Form.Get("TopicArn") != arn {
							t.Fatal("wrong absence target")
						}
						status = 404
						body = "<ErrorResponse><Error><Code>NotFound</Code><Message>missing</Message></Error></ErrorResponse>"
						switch mode {
						case "exists":
							status = 200
							body = "<GetTopicAttributesResponse/>"
						case "denied":
							status = 403
						case "ambiguous absence":
							body = "<ErrorResponse><Error><Code>NotFound</Code><Code>NotFound</Code></Error></ErrorResponse>"
						case "wrong absence code":
							body = strings.Replace(body, "NotFound", "NoSuchKey", 1)
						}
					} else if calls == 2 {
						if r.Form.Get("Action") != "CreateTopic" || r.Form.Get("Name") != "events" || r.Form.Get("push-endpoint") != "amqps://u:private-password@host/vhost" || r.Form.Get("time_to_live") != "-1" || r.Form.Get("max_retries") != "0" || r.Form.Get("retry_sleep_duration") != "10" || r.Form.Get("verify-ssl") != "true" || r.Form.Get("persistent") != "true" || r.Form.Get("Policy") != "{}" || r.Form.Get("OpaqueData") != "note &+" {
							t.Fatal("wrong creation configuration")
						}
						returned := arn
						if mode == "wrong returned ARN" {
							returned = "arn:wrong"
						}
						body = "<CreateTopicResponse><CreateTopicResult><TopicArn>" + returned + "</TopicArn></CreateTopicResult></CreateTopicResponse>"
						if mode == "create failure" {
							status = 500
							body = "private-password"
						}
					} else {
						if calls != 3 || r.Form.Get("Action") != "GetTopicAttributes" || r.Form.Get("TopicArn") != arn {
							t.Fatal("wrong verification")
						}
						dest := map[string]any{"EndpointAddress": "amqps://u:private-password@host/vhost", "EndpointArgs": "Version=2010-03-31&verify-ssl=true", "EndpointTopic": "events", "HasStoredSecret": true, "Persistent": true, "TimeToLive": "None", "MaxRetries": "0", "RetrySleepDuration": "10"}
						if mode == "wrong field" {
							dest["TimeToLive"] = "0"
						}
						if mode == "wrong args" {
							dest["EndpointArgs"] = ""
						}
						if mode == "wrong secret flag" {
							dest["HasStoredSecret"] = false
						}
						encoded, _ := json.Marshal(dest)
						attrs := map[string]string{"User": owner, "Name": "events", "TopicArn": arn, "OpaqueData": "note &+", "Policy": "{}", "EndPoint": string(encoded)}
						if mode == "wrong owner readback" {
							attrs["User"] = "other"
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
						if mode == "post denied" {
							status = 403
							body = "private-password"
						}
					}
					return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_topic.create", ResourceKey: "rgw/topic/" + id, Parameters: map[string]any{"topic_id": id, "topic_arn": arn, "owner_uid": "user", "endpoint_secret": "amqps://u:private-password@host/vhost", "opaque_data": "note &+", "policy": "{}", "persistent": true, "time_to_live": "None", "max_retries": "0", "retry_sleep_duration": "10", "options": map[string]any{"verify-ssl": "true"}}})
				code := "post_check_failed"
				wantCalls := 3
				switch mode {
				case "success":
					code = ""
				case "wrong owner":
					code = "pre_check_failed"
					wantCalls = 0
				case "exists", "denied", "ambiguous absence", "wrong absence code":
					code = "pre_check_failed"
					wantCalls = 1
				case "create failure", "wrong returned ARN":
					code = "sns_failed"
					wantCalls = 2
				}
				if calls != wantCalls || owners != 1 {
					t.Fatalf("calls %d owners %d", calls, owners)
				}
				if code == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || strings.Contains(err.Error(), "private-password") {
						t.Fatalf("unsafe result %v", err)
					}
				}
			})
		}
	}
}
