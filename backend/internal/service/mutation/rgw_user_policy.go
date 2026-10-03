package mutation

import "encoding/json"

func rgwUserPolicyMatches(raw []byte, parameters map[string]any) bool {
	var policies []string
	if json.Unmarshal(raw, &policies) != nil || policies == nil {
		return false
	}
	arn, ok := parameters["policy_arn"].(string)
	if !ok || arn == "" {
		return false
	}
	found := false
	seen := make(map[string]bool, len(policies))
	for _, policy := range policies {
		if policy == "" || seen[policy] {
			return false
		}
		seen[policy] = true
		if policy == arn {
			found = true
		}
	}
	switch parameters["action"] {
	case "attach":
		return found
	case "detach":
		return !found
	default:
		return false
	}
}
