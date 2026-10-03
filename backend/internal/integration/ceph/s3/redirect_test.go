package s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestSignedRequestsNeverFollowRedirects(t *testing.T) {
	for _, status := range []int{301, 302, 303, 307, 308} {
		for _, target := range []string{"https://other.example.test/capture", "http://s3.example.test/downgrade", "https://s3.example.test/other", "/relative"} {
			for _, operation := range []string{"s3-read", "s3-write", "s3-delete", "sns-write", "mfa-write"} {
				t.Run(operation+"/"+http.StatusText(status)+"/"+target, func(t *testing.T) {
					calls, redirectCalls := 0, 0
					transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
						calls++
						if calls != 1 {
							t.Fatal("signed request replayed to redirect target")
						}
						if r.URL.Host != "s3.example.test" || r.URL.Scheme != "https" || r.Header.Get("X-Amz-Security-Token") != "session-secret" || r.Header.Get("Authorization") == "" {
							t.Fatal("original request lost identity or credentials")
						}
						if operation == "sns-write" {
							data, _ := io.ReadAll(r.Body)
							form, _ := url.ParseQuery(string(data))
							if form.Get("AttributeValue") != "https://user:delivery-secret@push.example.test" {
								t.Fatal("SNS body not tested with a secret")
							}
						}
						return &http.Response{StatusCode: status, Header: http.Header{"Location": {target}}, Body: io.NopCloser(strings.NewReader(`<Error><Code>Redirect</Code></Error>`))}, nil
					})
					shared := &http.Client{Timeout: 7 * time.Second, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { redirectCalls++; return nil }}
					client, err := New("https://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret", SessionToken: "session-secret"}, shared)
					if err != nil {
						t.Fatal(err)
					}
					if client.http == shared || client.http.Timeout != shared.Timeout {
						t.Fatal("shared client or timeout changed")
					}
					switch operation {
					case "s3-read":
						_, _, err = client.GetBucketConfiguration(context.Background(), ":bucket", "notification")
					case "s3-write":
						err = client.PutBucketConfiguration(context.Background(), ":bucket", "versioning", []byte(`<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>`))
					case "s3-delete":
						err = client.DeleteBucketNotification(context.Background(), ":bucket", "single", "id")
					case "sns-write":
						err = client.SetTopicEndpoint(context.Background(), "arn:aws:sns:east::topic", "https://user:delivery-secret@push.example.test")
					case "mfa-write":
						err = client.PutBucketMFA(context.Background(), ":bucket", "Enabled", "Enabled", "serial", "001234")
					}
					if err == nil || calls != 1 || redirectCalls != 0 {
						t.Fatalf("redirect accepted/replayed: calls=%d redirect=%d err=%v", calls, redirectCalls, err)
					}
					if strings.HasPrefix(operation, "s3-") {
						var response *ResponseError
						if !errors.As(err, &response) || response.StatusCode != status {
							t.Fatal("redirect status lost")
						}
					}
					if strings.Contains(err.Error(), "delivery-secret") || strings.Contains(err.Error(), "session-secret") || strings.Contains(err.Error(), "001234") {
						t.Fatal("credential exposed in error")
					}
					_ = shared.CheckRedirect(nil, nil)
					if redirectCalls != 1 {
						t.Fatal("caller redirect policy mutated")
					}
				})
			}
		}
	}
}
