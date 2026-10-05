package mutation

import (
	"encoding/json"
	"strconv"
)

func rgwAccountUpdateMatches(parameters map[string]any, raw []byte) bool {
	var account map[string]json.RawMessage
	if json.Unmarshal(raw, &account) != nil {
		return false
	}
	var id *string
	if json.Unmarshal(account["id"], &id) != nil || id == nil || *id == "" || *id != rawText(parameters, "account_id") {
		return false
	}
	fields := 0
	for _, field := range []string{"account_name", "email"} {
		expected, supplied := parameters[field]
		if !supplied {
			continue
		}
		fields++
		key := field
		if key == "account_name" {
			key = "name"
		}
		want, ok := expected.(string)
		var actual *string
		if !ok || json.Unmarshal(account[key], &actual) != nil || actual == nil || *actual != want {
			return false
		}
	}
	for _, field := range []string{"max_users", "max_roles", "max_groups", "max_buckets", "max_access_keys"} {
		if _, supplied := parameters[field]; !supplied {
			continue
		}
		fields++
		want, err := strconv.ParseInt(optional(parameters, field), 10, 32)
		var actual *int32
		if err != nil || want < -1 || json.Unmarshal(account[field], &actual) != nil || actual == nil || int64(*actual) != want {
			return false
		}
	}
	return fields > 0
}
