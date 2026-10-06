package mutation

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

func osdIndividualFlagMatches(request Request, raw []byte) bool {
	var dump struct {
		OSDs []struct {
			ID    *int     `json:"osd"`
			State []string `json:"state"`
		} `json:"osds"`
	}
	if json.Unmarshal(raw, &dump) != nil || dump.OSDs == nil {
		return false
	}
	id := last(resourceTail(request.ResourceKey))
	flag, _ := request.Parameters["flag"].(string)
	action, _ := request.Parameters["action"].(string)
	found, matches := false, false
	for _, osd := range dump.OSDs {
		if osd.ID == nil {
			return false
		}
		if strconv.Itoa(*osd.ID) != id {
			continue
		}
		if found || osd.State == nil {
			return false
		}
		seen := map[string]bool{}
		for _, state := range osd.State {
			if state == "" || strings.TrimSpace(state) != state || seen[state] {
				return false
			}
			seen[state] = true
		}
		found = true
		matches = slices.Contains(osd.State, flag) == (action == "set")
	}
	return found && matches
}
