package mutation

import (
	"encoding/json"
	"strings"
)

func osdGlobalFlagMatches(parameters map[string]any, raw []byte) bool {
	var dump struct {
		Flags *string `json:"flags"`
	}
	if json.Unmarshal(raw, &dump) != nil || dump.Flags == nil {
		return false
	}
	flags := map[string]bool{}
	if *dump.Flags != "" {
		for _, flag := range strings.Split(*dump.Flags, ",") {
			if flag == "" || strings.TrimSpace(flag) != flag || flags[flag] {
				return false
			}
			flags[flag] = true
		}
	}
	flag, ok := parameters["flag"].(string)
	action, _ := parameters["action"].(string)
	if !ok || flag == "" || (action != "set" && action != "unset") {
		return false
	}
	if flag == "pause" {
		if action == "set" {
			return flags["pauserd"] && flags["pausewr"]
		}
		return !flags["pauserd"] && !flags["pausewr"]
	}
	return flags[flag] == (action == "set")
}
