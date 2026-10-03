package s3

import (
	"fmt"
	"strings"
	"testing"
)

func TestConfigurationMissingRequiresCompleteNativeError(t *testing.T) {
	for kind, code := range map[string]string{"policy": "NoSuchBucketPolicy", "cors": "NoSuchCORSConfiguration", "lifecycle": "NoSuchLifecycleConfiguration", "encryption": "ServerSideEncryptionConfigurationNotFoundError", "tagging": "NoSuchTagSet"} {
		body := "<Error><Code>" + code + "</Code><Message>missing</Message></Error>"
		if !IsConfigurationMissing(kind, fmt.Errorf("wrapped: %w", responseError(404, strings.NewReader(body)))) {
			t.Fatal(kind)
		}
		for _, invalid := range []string{
			body + "<Error/>", body + "garbage", strings.TrimSuffix(body, "</Error>"),
			"<Error><Code>" + code + "</Code><Code>" + code + "</Code></Error>",
			"<Error><Code><Nested/>" + code + "</Code></Error>",
			"<Other><Code>" + code + "</Code></Other>",
			"<Error><Code>NoSuchBucket</Code></Error>",
			body + strings.Repeat(" ", 32<<10),
		} {
			if IsConfigurationMissing(kind, responseError(404, strings.NewReader(invalid))) {
				t.Fatalf("accepted invalid envelope: %.100s", invalid)
			}
		}
		for _, status := range []int{403, 500} {
			if IsConfigurationMissing(kind, responseError(status, strings.NewReader(body))) {
				t.Fatal(status)
			}
		}
		for _, other := range []string{"unknown", "versioning", "policy", "cors", "lifecycle", "encryption", "tagging"} {
			if other != kind && IsConfigurationMissing(other, responseError(404, strings.NewReader(body))) {
				t.Fatal(other)
			}
		}
	}
}
