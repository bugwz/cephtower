package mutation

import (
	"encoding/json"
	"strconv"
)

// Native ratelimit get returns a scope wrapper, without an owner identity.
// Identity is supplied by the exact selectors in the readback command.
func rgwRateLimitMatches(action string, parameters map[string]any, raw []byte) bool {
	field := "user_ratelimit"
	if action == "rgw_bucket.ratelimit" {
		field = "bucket_ratelimit"
	} else if action != "rgw_user.ratelimit" {
		return false
	}
	var response map[string]json.RawMessage
	if json.Unmarshal(raw, &response) != nil {
		return false
	}
	var limits map[string]json.RawMessage
	if json.Unmarshal(response[field], &limits) != nil {
		return false
	}
	var enabled *bool
	wantEnabled, ok := parameters["enabled"].(bool)
	if !ok || json.Unmarshal(limits["enabled"], &enabled) != nil || enabled == nil || *enabled != wantEnabled {
		return false
	}
	for _, key := range []string{"max_read_ops", "max_write_ops", "max_read_bytes", "max_write_bytes"} {
		want, err := strconv.ParseInt(optional(parameters, key), 10, 64)
		var actual *int64
		if err != nil || want < 0 || want > 9007199254740991 || json.Unmarshal(limits[key], &actual) != nil || actual == nil || *actual != want {
			return false
		}
	}
	return true
}
