package ceph

import (
	"encoding/json"
	"strconv"
)

// Preserve the native unsigned counts before JSON reaches JavaScript consumers.
func enrichMirrorStateCounts(value any) {
	summary, ok := value.(map[string]any)
	if !ok {
		return
	}
	states, ok := summary["states"].(map[string]any)
	if !ok {
		return
	}
	for state, raw := range states {
		number, ok := raw.(json.Number)
		if !ok {
			states[state] = nil
			continue
		}
		if _, err := strconv.ParseUint(number.String(), 10, 64); err != nil {
			states[state] = nil
			continue
		}
		states[state] = number.String()
	}
}
