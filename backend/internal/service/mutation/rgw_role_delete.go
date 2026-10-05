package mutation

import (
	"encoding/json"
	"strings"
)

func rgwRoleDeleteName(parameters map[string]any) (string, error) {
	name, ok := parameters["name"].(string)
	if !ok || name != strings.TrimSpace(name) || len(name) > 512 {
		return "", invalid("name is required or invalid")
	}
	parts := strings.Split(name, "$")
	if len(parts) > 2 {
		return "", invalid("name is required or invalid")
	}
	for _, part := range parts {
		if !identifier.MatchString(part) {
			return "", invalid("name is required or invalid")
		}
	}
	return name, nil
}

func rgwRoleAbsent(parameters map[string]any, raw []byte) bool {
	var roles []struct {
		Name    *string `json:"RoleName"`
		Account *string `json:"AccountId"`
	}
	if json.Unmarshal(raw, &roles) != nil || roles == nil {
		return false
	}
	name, account := optional(parameters, "name"), optional(parameters, "account_id")
	if name == "" {
		return false
	}
	seen := map[string]bool{}
	for _, role := range roles {
		if role.Name == nil || *role.Name == "" || role.Account == nil || *role.Account != account || seen[*role.Name] || *role.Name == name {
			return false
		}
		seen[*role.Name] = true
	}
	return true
}
