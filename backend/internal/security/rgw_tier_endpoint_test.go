package security

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTierEndpointRedaction(t *testing.T) {
	for _, tc := range []struct {
		input  any
		want   string
		hidden bool
	}{
		{"https://user:private-pass@host.test:8443/path?a=private-query#private-fragment", "https://host.test:8443/path", true},
		{"https://host.test/path?X-Amz-Signature=private-signature", "https://host.test/path", true},
		{"http://[::1]:8080/p%2Fname", "http://[::1]:8080/p%2Fname", false},
		{"https://host.test/path?", "https://host.test/path", true},
		{"", "", false}, {nil, "[REDACTED]", true},
		{"javascript:private-value", "[REDACTED]", true},
		{"https://user:private-value@", "[REDACTED]", true},
		{"https://host.test/%zz?x=private-value", "[REDACTED]", true},
		{"https://host.test\\private-value", "[REDACTED]", true},
		{"https://host.test/\nprivate-value", "[REDACTED]", true},
	} {
		got, hidden := redactTierEndpoint(tc.input)
		if got != tc.want || hidden != tc.hidden {
			t.Fatalf("unexpected safe endpoint %q hidden=%v", got, hidden)
		}
	}
}

func TestTierEndpointRedactionNestedAndIdempotent(t *testing.T) {
	input := map[string]any{"current_period_details": map[string]any{"period_map": map[string]any{"zonegroups": []any{map[string]any{"placement_targets": []any{map[string]any{"name": "p", "tier_targets": []any{map[string]any{"key": "COLD", "val": map[string]any{"tier_type": "cloud-s3", "s3": map[string]any{"endpoint": "https://u:private-pass@host.test/path?custom=private-query#private-fragment", "secret": "private-secret", "access_key": "private-access", "region": "region", "multipart_min_part_size": 0}}}}}}}}}}}
	visible, err := RedactJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(visible)
	if strings.Contains(string(encoded), "private-") || !strings.Contains(string(encoded), `"endpoint":"https://host.test/path"`) || !strings.Contains(string(encoded), `"endpoint_redacted":true`) {
		t.Fatal("nested endpoint projection failed")
	}
	again, err := RedactJSON(visible)
	if err != nil || !reflect.DeepEqual(visible, again) {
		t.Fatal("redaction changed on second pass")
	}
	original, _ := json.Marshal(input)
	if !strings.Contains(string(original), "private-query") {
		t.Fatal("modified source")
	}
}
