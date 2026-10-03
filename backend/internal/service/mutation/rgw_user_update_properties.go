package mutation

import (
	"encoding/json"
	"strconv"
)

func rgwUserUpdatePropertiesMatch(raw []byte, p map[string]any, uid string) bool {
	requested := false
	for _, field := range []string{"display_name", "email", "max_buckets", "system", "suspended"} {
		if _, exists := p[field]; exists {
			requested = true
		}
	}
	if !requested {
		return true
	}
	var info struct {
		UID        string  `json:"full_user_id"`
		Name       *string `json:"display_name"`
		Email      *string `json:"email"`
		MaxBuckets *int64  `json:"max_buckets"`
		System     *bool   `json:"system"`
		Suspended  *int    `json:"suspended"`
	}
	if json.Unmarshal(raw, &info) != nil || info.UID != uid {
		return false
	}
	if name := rawText(p, "display_name"); name != "" && (info.Name == nil || *info.Name != name) {
		return false
	}
	if email, exists := p["email"].(string); exists && (info.Email == nil || *info.Email != email) {
		return false
	}
	if _, exists := p["max_buckets"]; exists {
		limit, err := strconv.ParseInt(optional(p, "max_buckets"), 10, 32)
		if err != nil || info.MaxBuckets == nil || *info.MaxBuckets != limit {
			return false
		}
	}
	if system, exists := p["system"].(bool); exists && (info.System == nil || *info.System != system) {
		return false
	}
	if suspended, exists := p["suspended"].(bool); exists {
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
