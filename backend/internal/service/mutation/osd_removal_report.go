package mutation

import (
	"encoding/json"
	"strconv"
)

// A successful JSON command can still report unsafe OSDs; never infer safety
// from its exit status alone (DaemonServer's safe-to-destroy handler).
func osdRemovalReport(raw []byte, ids []string) (map[string]any, bool) {
	var fields map[string][]*int
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, false
	}
	wanted := map[int]bool{}
	for _, id := range ids {
		n, err := strconv.Atoi(id)
		if err != nil {
			return nil, false
		}
		wanted[n] = true
	}
	report := map[string]any{}
	sets := map[string]map[int]bool{}
	classified := map[int]bool{}
	for _, key := range []string{"safe_to_destroy", "active", "missing_stats", "stored_pgs"} {
		values := fields[key]
		if values == nil {
			return nil, false
		}
		set := map[int]bool{}
		out := make([]int, 0, len(values))
		for _, value := range values {
			if value == nil || !wanted[*value] || classified[*value] {
				return nil, false
			}
			set[*value] = true
			classified[*value] = true
			out = append(out, *value)
		}
		sets[key] = set
		report[key] = out
	}
	report["is_safe_to_destroy"] = len(wanted) > 0 && len(sets["safe_to_destroy"]) == len(wanted)
	return report, true
}
