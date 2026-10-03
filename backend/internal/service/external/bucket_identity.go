package external

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Match the resource key emitted by the API, never infer identity from its tail.
func bucketResourceID(action, key string) (string, error) {
	parts := strings.Split(key, "/")
	valid := len(parts) == 3 && (action == "rgw_bucket.update" || action == "rgw_bucket.delete" || action == "rgw_bucket.acl" || action == "rgw_bucket.replication_enable" || action == "rgw_bucket.notification_delete" || action == "rgw_bucket.notification_set")
	if action == "rgw_bucket_policy.update" || action == "rgw_bucket_policy.delete" {
		valid = len(parts) == 4 && parts[3] == "policy"
	}
	if !valid || parts[0] != "rgw" || parts[1] != "bucket" || parts[2] == "" {
		return "", fmt.Errorf("invalid bucket resource key")
	}
	return parts[2], nil
}

func decodeBucketID(value string) (string, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || !utf8.Valid(decoded) || base64.RawURLEncoding.EncodeToString(decoded) != value {
		return "", fmt.Errorf("invalid bucket ID")
	}
	parts := strings.Split(string(decoded), "\x00")
	if len(parts) != 2 || parts[1] == "" {
		return "", fmt.Errorf("bucket ID must encode tenant and bucket")
	}
	for _, part := range parts {
		if strings.ContainsAny(part, "/:\\") || strings.IndexFunc(part, func(r rune) bool { return unicode.IsControl(r) || unicode.IsSpace(r) }) >= 0 {
			return "", fmt.Errorf("invalid bucket identity component")
		}
	}
	// rgw_parse_url_bucket treats bare names as belonging to the credential's
	// tenant. Even the global tenant must be explicit (":bucket").
	return parts[0] + ":" + parts[1], nil
}
