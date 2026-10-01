package mutation

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func applySMBLoginControl(record, parameters map[string]any) error {
	rules, hasRules := parameters["login_control"]
	restricted, hasRestricted := parameters["restrict_access"]
	if !hasRules && !hasRestricted {
		return nil
	}
	if !hasRules || !hasRestricted {
		return invalid("login_control and restrict_access must be replaced together")
	}
	flag, ok := restricted.(bool)
	if !ok || rules == nil {
		return invalid("invalid SMB login control configuration")
	}
	raw, err := json.Marshal(rules)
	var entries []map[string]any
	if err != nil || json.Unmarshal(raw, &entries) != nil || entries == nil {
		return invalid("login_control must be an array")
	}
	seen := map[string]bool{}
	allowed := false
	for _, entry := range entries {
		name, _ := entry["name"].(string)
		category, _ := entry["category"].(string)
		access, _ := entry["access"].(string)
		if name == "" || utf8.RuneCountInString(name) > 128 || strings.ContainsAny(name, " \t\r\n\x00") || (category != "user" && category != "group") || (access != "none" && access != "read" && access != "read-write" && access != "admin") || len(entry) != 3 || seen[category+"\x00"+name] {
			return invalid("invalid or duplicate SMB login control entry")
		}
		seen[category+"\x00"+name] = true
		allowed = allowed || access != "none"
	}
	if flag && !allowed {
		return invalid("restricted SMB access requires at least one allowed user or group")
	}
	record["login_control"] = entries
	record["restrict_access"] = flag
	return nil
}
