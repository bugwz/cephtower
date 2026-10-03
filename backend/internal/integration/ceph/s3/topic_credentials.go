package s3

import (
	"net/url"
	"strings"
)

type TopicCredentialState struct {
	Field string `json:"field"`
	State string `json:"state"`
}

// TopicCredentialPresence reports only raw argument presence, never values or
// effective authentication. RGW decodes each segment before splitting its key.
// The stored_secret marker can be stale after SetTopicAttributes.
func TopicCredentialPresence(raw string) []TopicCredentialState {
	result := map[string]string{"username": "unset", "password": "unset"}
	counts := map[string]int{}
	for _, segment := range strings.Split(strings.TrimPrefix(raw, "?"), "&") {
		decoded, err := url.QueryUnescape(segment)
		if err != nil {
			return []TopicCredentialState{{"username", "unavailable"}, {"password", "unavailable"}}
		}
		name, value, _ := strings.Cut(decoded, "=")
		key := ""
		switch name {
		case "user-name":
			key = "username"
		case "password":
			key = "password"
		default:
			continue
		}
		counts[key]++
		switch {
		case counts[key] > 1:
			result[key] = "duplicate"
		case value == "":
			result[key] = "empty"
		default:
			result[key] = "set"
		}
	}
	return []TopicCredentialState{{"username", result["username"]}, {"password", result["password"]}}
}
