package mutation

import "encoding/json"

func rgwRoleCreateMatches(parameters map[string]any, raw []byte) bool {
	requested := make(map[string]any, len(parameters)+1)
	for key, value := range parameters {
		requested[key] = value
	}
	if _, present := requested["max_session_duration"]; !present {
		requested["max_session_duration"] = json.Number("3600")
	}
	if !rgwRoleUpdateMatches(requested, raw) {
		return false
	}
	var role struct {
		ID          *string `json:"RoleId"`
		Path        *string `json:"Path"`
		Description *string `json:"Description"`
	}
	path := rawText(parameters, "path")
	if path == "" {
		path = "/"
	}
	return json.Unmarshal(raw, &role) == nil && role.ID != nil && *role.ID != "" && role.Path != nil && *role.Path == path && role.Description != nil && *role.Description == rawText(parameters, "description")
}
