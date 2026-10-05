package mutation

import "encoding/json"

func rgwAccountCreateMatches(parameters map[string]any, raw []byte) bool {
	// Native create initializes these text fields from empty option strings.
	// Limits not explicitly supplied are server defaults and are not guessed.
	expected := make(map[string]any, len(parameters)+2)
	for key, value := range parameters {
		expected[key] = value
	}
	for _, key := range []string{"account_name", "email"} {
		if _, exists := expected[key]; !exists {
			expected[key] = ""
		}
	}
	if !rgwAccountUpdateMatches(expected, raw) {
		return false
	}
	var account struct {
		Tenant *string `json:"tenant"`
	}
	if json.Unmarshal(raw, &account) != nil || account.Tenant == nil {
		return false
	}
	tenant := ""
	if value, exists := parameters["tenant"]; exists {
		var ok bool
		tenant, ok = value.(string)
		if !ok {
			return false
		}
	}
	return *account.Tenant == tenant
}
