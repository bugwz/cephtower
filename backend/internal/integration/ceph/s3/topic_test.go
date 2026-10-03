package s3

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSNSPolicyTransportGuards(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, body string
		status               int
		calls                int
	}{
		{"https required", "http://example.test", "", 200, 0},
		{"wrong root", "https://example.test", "<Other/>", 200, 1},
		{"trailing document", "https://example.test", "<SetTopicAttributesResponse/><Other/>", 200, 1},
		{"remote failure", "https://example.test", "sensitive-endpoint-password", 403, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client, err := New(tc.endpoint, Credentials{AccessKey: "key", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				body, _ := io.ReadAll(r.Body)
				if r.Header.Get("X-Amz-Content-Sha256") != sha256Hex(body) {
					t.Fatal("wrong payload hash")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: http.Header{}}, nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			err = client.SetTopicPolicy(context.Background(), "arn:aws:sns:default::events", "")
			if err == nil || strings.Contains(err.Error(), "sensitive-endpoint-password") || calls != tc.calls {
				t.Fatalf("unsafe outcome %v calls=%d", err, calls)
			}
		})
	}
}

func TestSNSAttributesRejectAmbiguity(t *testing.T) {
	entries := ""
	for _, key := range []string{"User", "Name", "EndPoint", "TopicArn", "OpaqueData", "Policy"} {
		entries += "<entry><key>" + key + "</key><value></value></entry>"
	}
	valid := "<GetTopicAttributesResponse><GetTopicAttributesResult><Attributes>" + entries + "</Attributes></GetTopicAttributesResult><ResponseMetadata/></GetTopicAttributesResponse>"
	attrs, err := parseTopicAttributes([]byte(valid))
	if err != nil || len(attrs) != 6 || attrs["Policy"] != "" {
		t.Fatalf("empty attributes: %v", err)
	}
	for _, body := range []string{
		strings.Replace(valid, "<key>User</key>", "<key>Policy</key>", 1),
		strings.Replace(valid, "<value></value>", "<value/><value/>", 1),
		strings.Replace(valid, "<value></value>", "<value><nested/></value>", 1),
		strings.Replace(valid, "<key>User</key>", "<key>Other</key>", 1),
		valid + "<Other/>", "<GetTopicAttributesResponse/>",
	} {
		if _, err := parseTopicAttributes([]byte(body)); err == nil {
			t.Fatal("ambiguous response accepted")
		}
	}
}
