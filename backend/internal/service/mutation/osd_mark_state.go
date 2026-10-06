package mutation

import (
	"encoding/json"
	"strconv"
)

func osdMarkStateMatches(request Request, raw []byte) bool {
	var dump struct {
		OSDs []struct {
			ID *int `json:"osd"`
			Up *int `json:"up"`
			In *int `json:"in"`
		} `json:"osds"`
	}
	if json.Unmarshal(raw, &dump) != nil || dump.OSDs == nil {
		return false
	}
	id := pathValue(resourceTail(request.ResourceKey), "osd")
	action, _ := request.Parameters["action"].(string)
	found, matches := false, false
	for _, osd := range dump.OSDs {
		if osd.ID == nil {
			return false
		}
		if strconv.Itoa(*osd.ID) != id {
			continue
		}
		if found {
			return false
		}
		found = true
		switch action {
		case "in":
			matches = osd.In != nil && *osd.In == 1
		case "out":
			matches = osd.In != nil && *osd.In == 0
		case "down":
			matches = osd.Up != nil && *osd.Up == 0
		default:
			return false
		}
	}
	return found && matches
}
