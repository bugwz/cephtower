package mutation

import (
	"encoding/json"
	"strconv"
)

// user info identifies users with full_user_id, while account get uses id.
// Quota scopes are independent: never verify against another scope's limits.
func rgwOwnerQuotaMatches(action string, parameters map[string]any, data []byte) bool {
	var actual struct {
		UID          *string         `json:"full_user_id"`
		AccountID    *string         `json:"id"`
		UserQuota    json.RawMessage `json:"user_quota"`
		AccountQuota json.RawMessage `json:"quota"`
		BucketQuota  json.RawMessage `json:"bucket_quota"`
	}
	if json.Unmarshal(data, &actual) != nil {
		return false
	}
	var raw json.RawMessage
	scope := rawText(parameters, "scope")
	switch action {
	case "rgw_user.quota":
		uid := rawText(parameters, "uid")
		if uid == "" || actual.UID == nil || *actual.UID != uid || (scope != "user" && scope != "bucket") {
			return false
		}
		raw = actual.UserQuota
	case "rgw_account.quota":
		id := rawText(parameters, "account_id")
		if id == "" || actual.AccountID == nil || *actual.AccountID != id || (scope != "account" && scope != "bucket") {
			return false
		}
		raw = actual.AccountQuota
	default:
		return false
	}
	if scope == "bucket" {
		raw = actual.BucketQuota
	}
	var quota struct {
		Enabled    *bool  `json:"enabled"`
		MaxSize    *int64 `json:"max_size"`
		MaxObjects *int64 `json:"max_objects"`
	}
	if json.Unmarshal(raw, &quota) != nil || quota.Enabled == nil || quota.MaxSize == nil || quota.MaxObjects == nil {
		return false
	}
	enabled, ok := parameters["enabled"].(bool)
	size, sizeErr := strconv.ParseInt(optional(parameters, "max_size"), 10, 64)
	objects, objectsErr := strconv.ParseInt(optional(parameters, "max_objects"), 10, 64)
	if !ok || sizeErr != nil || objectsErr != nil || size < -1 || size > 9007199254740991 || objects < -1 || objects > 9007199254740991 {
		return false
	}
	if size >= 0 {
		size = (size + 1023) / 1024 * 1024
	}
	return *quota.Enabled == enabled && *quota.MaxSize == size && *quota.MaxObjects == objects
}
