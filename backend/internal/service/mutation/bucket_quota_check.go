package mutation

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

// The native bucket quota dispatcher can return success after a failed write.
// Require the subsequent stats to contain the requested identity and quota.
func bucketQuotaMatches(parameters map[string]any, data []byte) bool {
	var actual struct {
		Bucket string `json:"bucket"`
		Tenant string `json:"tenant"`
		Quota  *struct {
			Enabled    *bool  `json:"enabled"`
			MaxSize    *int64 `json:"max_size"`
			MaxObjects *int64 `json:"max_objects"`
		} `json:"bucket_quota"`
	}
	if json.Unmarshal(data, &actual) != nil || actual.Quota == nil || actual.Quota.Enabled == nil || actual.Quota.MaxSize == nil || actual.Quota.MaxObjects == nil {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(rawText(parameters, "bucket_id"))
	parts := strings.SplitN(string(raw), "\x00", 2)
	if err != nil || len(parts) != 2 || actual.Tenant != parts[0] || actual.Bucket != parts[1] {
		return false
	}
	enabled, ok := parameters["enabled"].(bool)
	if !ok || *actual.Quota.Enabled != enabled {
		return false
	}
	size, err := strconv.ParseInt(optional(parameters, "max_size"), 10, 64)
	if err != nil || size < -1 || size > 9007199254740991 {
		return false
	}
	if size >= 0 {
		size = ((size + 1023) / 1024) * 1024
	}
	objects, err := strconv.ParseInt(optional(parameters, "max_objects"), 10, 64)
	return err == nil && *actual.Quota.MaxSize == size && *actual.Quota.MaxObjects == objects
}
