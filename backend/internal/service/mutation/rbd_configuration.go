package mutation

import "encoding/json"

// Native config image list returns effective values together with their scope.
// A matching inherited value does not prove that an image override was set.
func rbdImageConfigurationMatches(parameters map[string]any, data []byte) bool {
	var rows []struct {
		Name   string  `json:"name"`
		Value  *string `json:"value"`
		Source string  `json:"source"`
	}
	if json.Unmarshal(data, &rows) != nil || rows == nil {
		return false
	}
	name := optional(parameters, "config_name")
	found, matched := false, false
	seen := make(map[string]bool)
	for _, row := range rows {
		if row.Name == "" || row.Value == nil || row.Source == "" || seen[row.Name] {
			return false
		}
		seen[row.Name] = true
		if row.Name != name {
			continue
		}
		found = true
		switch optional(parameters, "action") {
		case "config-set":
			value, ok := parameters["config_value"].(string)
			matched = ok && row.Source == "image" && *row.Value == value
		case "config-remove":
			matched = row.Source == "config" || row.Source == "pool"
		}
	}
	return found && matched
}
