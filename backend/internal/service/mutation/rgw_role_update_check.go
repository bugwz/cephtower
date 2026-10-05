package mutation

import (
	"encoding/json"
	"strconv"
)

func rgwRoleUpdateMatches(parameters map[string]any, raw []byte) bool {
	var role struct {
		Name     *string `json:"RoleName"`
		Account  *string `json:"AccountId"`
		Trust    *string `json:"AssumeRolePolicyDocument"`
		Duration *int64  `json:"MaxSessionDuration"`
	}
	if json.Unmarshal(raw, &role) != nil || role.Name == nil || *role.Name != optional(parameters, "name") || role.Account == nil || *role.Account != optional(parameters, "account_id") {
		return false
	}
	changed := false
	if _, present := parameters["assume_role_policy"]; present {
		changed = true
		if role.Trust == nil || *role.Trust != rawText(parameters, "assume_role_policy") {
			return false
		}
	}
	if _, present := parameters["max_session_duration"]; present {
		changed = true
		duration, err := strconv.ParseInt(optional(parameters, "max_session_duration"), 10, 64)
		if err != nil || duration < 3600 || duration > 43200 || role.Duration == nil || *role.Duration != duration {
			return false
		}
	}
	return changed
}
