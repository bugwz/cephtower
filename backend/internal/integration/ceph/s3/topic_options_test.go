package s3

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTopicOptionsSafeProjection(t *testing.T) {
	raw := "verify-ssl=false&use-ssl=TRUE&cloudevents=true&ca-location=%2Fetc%2Fceph%2Fca.pem&amqp-version=0-9-1&amqp-exchange=events&amqp-ack-level=routable&http-ack-level=202&kafka-ack-level=broker&mechanism=SCRAM-SHA-256&kafka-brokers=a:9092,b:9093&user-name=never-store&password=never-store&unknown=never-store"
	options, ok := TopicOptions(raw)
	if !ok || len(options) != 11 {
		t.Fatal("valid projection missing")
	}
	for _, option := range options {
		if option.Status != "returned" || option.Value == nil {
			t.Fatalf("missing option %s", option.Name)
		}
	}
	body, _ := json.Marshal(options)
	if strings.Contains(string(body), "never-store") || strings.Contains(string(body), "password") || strings.Contains(string(body), "user-name") {
		t.Fatal("credential or unknown field leaked")
	}
	for _, tc := range []struct{ raw, name, status, value string }{
		{"", "verify-ssl", "unset", ""},
		{"verify-ssl=false&verify-ssl=true", "verify-ssl", "duplicate", ""},
		{"verify-ssl=never-store", "verify-ssl", "hidden_invalid", ""},
		{"verify-ssl%3Dfalse", "verify-ssl", "returned", "false"},
		{"?ca-location=/etc/my+ca.pem", "ca-location", "returned", "/etc/my ca.pem"},
		{"ca-location=", "ca-location", "returned", ""},
		{"amqp-exchange=", "amqp-exchange", "returned", ""},
		{"mechanism=never-store", "mechanism", "hidden_invalid", ""},
		{"kafka-brokers=user:never-store@broker:9092", "kafka-brokers", "hidden_invalid", ""},
		{"ca-location=https://user:never-store@host/file", "ca-location", "hidden_invalid", ""},
		{"amqp-exchange=events%26password%3Dnever-store", "amqp-exchange", "hidden_invalid", ""},
		{"http-ack-level=600", "http-ack-level", "hidden_invalid", ""},
		{"amqp-exchange=<script>never-store</script>", "amqp-exchange", "hidden_invalid", ""},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			options, ok := TopicOptions(tc.raw)
			if !ok {
				t.Fatal("unexpected parse failure")
			}
			found := false
			for _, option := range options {
				if option.Name == tc.name {
					found = true
					if option.Status != tc.status {
						t.Fatalf("status %s", option.Status)
					}
					if tc.status == "returned" {
						if option.Value == nil || *option.Value != tc.value {
							t.Fatal("value changed")
						}
					} else if option.Value != nil {
						t.Fatal("unsafe value retained")
					}
				}
			}
			if !found {
				t.Fatal("missing known option")
			}
			body, _ := json.Marshal(options)
			if strings.Contains(string(body), "never-store") {
				t.Fatal("secret leaked")
			}
		})
	}
	if options, ok := TopicOptions("verify-ssl=true&password=%zz"); ok || options != nil {
		t.Fatal("malformed input partially presented as complete")
	}
}
