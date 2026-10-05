package mutation

import "encoding/json"

func buildLocalStorageClassDelete(p map[string]any) (command, error) {
	if !syncFlowToken(syncGroupString(p, "zone_id")) || !syncFlowToken(syncGroupString(p, "zone_name")) || p["storage_class"] == "STANDARD" {
		return command{}, invalid("explicit Zone identity and non-STANDARD local class required")
	}
	return buildZonegroupStorageClassDelete(p)
}

func localStorageClassDeleteExpected(zone, p map[string]any) (map[string]any, bool) {
	raw, err := json.Marshal(zone)
	if err != nil {
		return nil, false
	}
	expected := periodDocument(raw)
	clear(raw)
	entries, ok := expected["placement_pools"].([]any)
	if !ok {
		return nil, false
	}
	found := false
	seen := map[string]bool{}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		key := syncGroupString(entry, "key")
		if !ok || !syncFlowToken(key) || seen[key] {
			return nil, false
		}
		seen[key] = true
		if key != p["placement_id"] {
			continue
		}
		info, ok := entry["val"].(map[string]any)
		if !ok {
			return nil, false
		}
		classes, ok := info["storage_classes"].(map[string]any)
		if !ok {
			return nil, false
		}
		if _, ok := classes["STANDARD"].(map[string]any); !ok {
			return nil, false
		}
		class := syncGroupString(p, "storage_class")
		if class == "STANDARD" {
			return nil, false
		}
		if _, ok := classes[class].(map[string]any); !ok {
			return nil, false
		}
		delete(classes, class)
		found = true
	}
	return expected, found
}
