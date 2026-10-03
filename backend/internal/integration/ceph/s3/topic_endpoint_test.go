package s3

import "testing"

func TestTopicEndpointValidationAndProjection(t *testing.T) {
	for _, value := range []string{"", "https://host/path", "http://host/path?q=1&x=2", "amqp://user:pass@host/vhost", "amqps://host:5671/", "kafka://host:9092", "https://host/path?q=secret#fragment"} {
		if _, err := ValidateTopicEndpoint(value); err != nil {
			t.Fatalf("valid endpoint rejected: %v", err)
		}
	}
	for _, value := range []string{"file:///tmp/a", "ftp://host/a", "https://user@host/a", "https://user:@host/a", "https://:pass@host/a", "https://host?query=1", "https://host/\n", "https://[::1]/", "https://host_with_underscore/path", "https://host/%zz"} {
		if _, err := ValidateTopicEndpoint(value); err == nil {
			t.Fatal("invalid native endpoint accepted")
		}
	}
	visible, hidden, valid := RedactTopicEndpoint("amqps://user:secret@host/vhost?q=private#private")
	if visible != "amqps://host/vhost" || !hidden || !valid {
		t.Fatal("endpoint secrets exposed")
	}
	visible, hidden, valid = RedactTopicEndpoint("not-a-url-secret")
	if visible != "" || !hidden || valid {
		t.Fatal("malformed endpoint exposed")
	}
}
