package mutation

import (
	"encoding/json"
	"math"
	"strconv"
)

func osdMarkStateMatches(request Request, raw []byte) bool {
	var dump struct {
		OSDs []struct {
			ID     *int     `json:"osd"`
			Up     *int     `json:"up"`
			In     *int     `json:"in"`
			Weight *float64 `json:"weight"`
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
		case "reweight":
			weight, err := strconv.ParseFloat(optional(request.Parameters, "weight"), 64)
			if err != nil || math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 || weight > 1 {
				return false
			}
			// OSDMonitor truncates CEPH_OSD_IN * weight to an integer; OSDMap
			// reports that integer divided by 0x10000, an exact binary fraction.
			expected := float64(uint32(weight*65536)) / 65536
			matches = osd.Weight != nil && *osd.Weight == expected
		default:
			return false
		}
	}
	return found && matches
}
