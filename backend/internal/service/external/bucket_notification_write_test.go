package external

import (
	"context"
	"encoding/base64"
	"encoding/xml"
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

func TestBucketNotificationWrites(t *testing.T) {
	service, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "s3", URL: "https://s3.example.test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := endpoints.PutCredential(ctx, cluster.ID, endpointservice.CredentialInput{Kind: "s3", Value: map[string]any{"access_key": "access", "secret_key": "secret"}}); err != nil {
		t.Fatal(err)
	}
	wrap := func(s string) string { return "<NotificationConfiguration>" + s + "</NotificationConfiguration>" }
	const old = `<TopicConfiguration><Id>id</Id><Topic>arn:aws:sns:east::old</Topic><Event>s3:ObjectCreated:*</Event></TopicConfiguration>`
	const other = `<TopicConfiguration><Id>other</Id><Topic>arn:aws:sns:east::other</Topic><Event>future</Event></TopicConfiguration>`
	for _, tenant := range []string{"", "team"} {
		for _, mode := range []string{"create", "edit", "move", "rescope"} {
			for _, scenario := range []string{"success", "empty", "snapshot", "pre_denied", "pre_broken", "topic_denied", "topic_identity", "put_failed", "post_failed", "post_changed", "duplicate", "delete_failed", "delete_unverified", "collision"} {
				if scenario == "empty" && mode != "create" {
					continue
				}
				if (scenario == "delete_failed" || scenario == "delete_unverified") && mode != "move" {
					continue
				}
				t.Run(tenant+"/"+mode+"/"+scenario, func(t *testing.T) {
					before, topic, selectedMode := wrap(other), "arn:aws:sns:east::new", "create"
					if mode != "create" {
						before, selectedMode = wrap(other+old), "edit"
					}
					if mode == "edit" {
						topic = "arn:aws:sns:east::old"
					}
					if mode == "rescope" {
						topic = "arn:aws:sns:east:RGW12345678901234567:old"
					}
					if scenario == "empty" {
						before = wrap("")
					}
					if scenario == "pre_broken" {
						before = "broken"
					}
					if scenario == "duplicate" {
						before = wrap(other + old + old)
					}
					if scenario == "collision" {
						before = wrap(`<TopicConfiguration><Id>id_new</Id><Topic>arn:aws:sns:east::topic</Topic></TopicConfiguration>`)
						topic = "arn:aws:sns:east::new_topic"
					}
					rule := s3.BucketNotification{ID: "id", Topic: topic, Events: []string{}, Filters: []s3.BucketNotificationFilter{{Kind: "S3Tags", Name: "key", Value: "<&"}}}
					parameters := map[string]any{"mode": selectedMode, "expected_document": before, "rule": rule}
					if scenario == "snapshot" {
						parameters["expected_document"] = "changed"
					}
					steps := []string{"read", "topic", "put", "post"}
					if mode == "move" {
						steps = []string{"read", "topic", "delete", "verify-delete", "put", "post"}
					}
					calls := 0
					service.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
						if calls >= len(steps) {
							t.Fatal("unexpected retry")
						}
						step := steps[calls]
						calls++
						body, status, method, path, query, signing := before, 200, "GET", "/"+tenant+":bucket", "notification=", "/s3/aws4_request"
						switch step {
						case "read":
							if scenario == "pre_denied" {
								status, body = 403, "secret-error"
							}
						case "topic":
							method, path, query, signing = "POST", "/", "", "/sns/aws4_request"
							data, _ := io.ReadAll(r.Body)
							form, _ := url.ParseQuery(string(data))
							if form.Get("Action") != "GetTopicAttributes" || form.Get("TopicArn") != topic {
								t.Fatal("wrong destination preflight")
							}
							var b strings.Builder
							b.WriteString("<GetTopicAttributesResponse><GetTopicAttributesResult><Attributes>")
							name, _ := s3.NotificationTopicName(topic)
							arn := topic
							if scenario == "topic_identity" {
								arn += "wrong"
							}
							for key, value := range map[string]string{"User": "owner", "Name": name, "TopicArn": arn, "EndPoint": `{}`, "Policy": "", "OpaqueData": ""} {
								b.WriteString("<entry><key>" + key + "</key><value>")
								_ = xml.EscapeText(&b, []byte(value))
								b.WriteString("</value></entry>")
							}
							b.WriteString("</Attributes></GetTopicAttributesResult></GetTopicAttributesResponse>")
							body = b.String()
							if scenario == "topic_denied" {
								status, body = 403, "secret-error"
							}
						case "delete":
							method, query, body = "DELETE", "notification=id", ""
							if scenario == "delete_failed" {
								status, body = 500, "secret-error"
							}
						case "verify-delete":
							body = wrap(other)
							if scenario == "delete_unverified" {
								body = before
							}
						case "put":
							method, body = "PUT", ""
							data, _ := io.ReadAll(r.Body)
							want, _, _ := s3.BucketNotificationDocument(rule)
							if string(data) != string(want) || r.Header.Get("Content-MD5") == "" {
								t.Fatal("wrong native body")
							}
							if scenario == "put_failed" {
								status, body = 500, "secret-error"
							}
						case "post":
							body = wrap(`<TopicConfiguration><Id>id</Id><Topic>` + topic + `</Topic><Event>s3:ObjectCreated:*</Event><Event>s3:ObjectRemoved:*</Event><Filter><S3Tags><FilterRule><Name>key</Name><Value>&lt;&amp;</Value></FilterRule></S3Tags></Filter></TopicConfiguration>` + other)
							if scenario == "empty" {
								body = strings.Replace(body, other, "", 1)
							}
							if scenario == "post_failed" {
								status, body = 404, `<Error><Code>NoSuchBucket</Code></Error>`
							}
							if scenario == "post_changed" {
								body = strings.Replace(body, "future", "changed", 1)
							}
						}
						if r.Method != method || r.URL.Path != path || r.URL.RawQuery != query || !strings.Contains(r.Header.Get("Authorization"), signing) {
							t.Fatalf("bad native request at %s: %s %s", step, r.Method, r.URL)
						}
						return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					id := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket"))
					_, err := service.Execute(ctx, Request{ClusterID: cluster.ID, Action: "rgw_bucket.notification_set", ResourceKey: "rgw/bucket/" + id, Parameters: parameters})
					wantCalls := len(steps)
					switch scenario {
					case "snapshot", "pre_denied", "pre_broken", "duplicate", "collision":
						wantCalls = 1
					case "topic_denied", "topic_identity":
						wantCalls = 2
					case "delete_failed":
						wantCalls = 3
					case "delete_unverified":
						wantCalls = 4
					case "put_failed":
						wantCalls--
					}
					if calls != wantCalls || (err == nil) != (scenario == "success" || scenario == "empty") {
						t.Fatalf("calls %d want %d, error %v", calls, wantCalls, err)
					}
					if err != nil {
						var action *cephdomain.ActionError
						if !errors.As(err, &action) || action.Retryable || strings.Contains(err.Error(), "secret-error") {
							t.Fatalf("unsafe error %v", err)
						}
					}
				})
			}
		}
	}
}
