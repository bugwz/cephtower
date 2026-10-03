package mutation

import (
	"encoding/json"
	"strconv"
)

// Compare values actually sent by user create, not unspecified cluster defaults.
func rgwUserCreatePropertiesMatch(raw []byte, p map[string]any) bool {
	var info struct {
		UID        string  `json:"full_user_id"`
		Name       *string `json:"display_name"`
		Email      *string `json:"email"`
		MaxBuckets *int64  `json:"max_buckets"`
		System     *bool   `json:"system"`
		Suspended  *int    `json:"suspended"`
	}
	if json.Unmarshal(raw, &info) != nil || info.UID != rawText(p, "uid") || info.Name == nil || *info.Name != rawText(p, "display_name") {
		return false
	}
	if email := rawText(p, "email"); email != "" && (info.Email == nil || *info.Email != email) {
		return false
	}
	if _, provided := p["max_buckets"]; provided {
		limit, err := strconv.ParseInt(optional(p, "max_buckets"), 10, 32)
		if err != nil || info.MaxBuckets == nil || *info.MaxBuckets != limit {
			return false
		}
	}
	if system, requested := p["system"].(bool); requested && (info.System == nil || *info.System != system) {
		return false
	}
	if suspended, requested := p["suspended"].(bool); requested {
		want := 0
		if suspended {
			want = 1
		}
		if info.Suspended == nil || *info.Suspended != want {
			return false
		}
	}
	return true
}
