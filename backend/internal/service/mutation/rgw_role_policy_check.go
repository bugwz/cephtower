package mutation

import "encoding/json"

func rgwRolePolicyMatches(parameters map[string]any, raw []byte) bool {
	var role struct {
		Name     *string         `json:"RoleName"`
		Account  *string         `json:"AccountId"`
		Policies json.RawMessage `json:"PermissionPolicies"`
	}
	if json.Unmarshal(raw, &role) != nil || role.Name == nil || *role.Name != optional(parameters, "name") || role.Account == nil || *role.Account != optional(parameters, "account_id") {
		return false
	}
	action, target := optional(parameters, "action"), optional(parameters, "policy_name")
	if target == "" || (action != "put" && action != "delete") {
		return false
	}
	if len(role.Policies) == 0 {
		return action == "delete"
	}
	var policies []struct {
		Name  *string `json:"PolicyName"`
		Value *string `json:"PolicyValue"`
	}
	if json.Unmarshal(role.Policies, &policies) != nil || policies == nil {
		return false
	}
	seen := map[string]bool{}
	found := false
	for _, policy := range policies {
		if policy.Name == nil || *policy.Name == "" || policy.Value == nil || seen[*policy.Name] {
			return false
		}
		seen[*policy.Name] = true
		if *policy.Name == target {
			found = true
			if action == "delete" || *policy.Value != rawText(parameters, "policy_document") {
				return false
			}
		}
	}
	return action == "delete" || found
}
