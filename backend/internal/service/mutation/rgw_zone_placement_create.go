package mutation

import (
	"encoding/json"
	"sort"
)

func zonePlacementCreateParams(p map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range p {
		result[key] = value
	}
	result["storage_class"] = "STANDARD"
	return result
}

func buildZonePlacementCreate(p map[string]any) (command, error) {
	spec, err := buildZonePlacement(zonePlacementCreateParams(p))
	if err != nil {
		return command{}, err
	}
	spec.args[2] = "add"
	return spec, nil
}

func zonePlacementCreateExpected(zone, p map[string]any) (map[string]any, bool) {
	raw, err := json.Marshal(zone)
	if err != nil {
		return nil, false
	}
	expected := periodDocument(raw)
	clear(raw)
	placements, ok := expected["placement_pools"].([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	for _, raw := range placements {
		entry, ok := raw.(map[string]any)
		key := syncGroupString(entry, "key")
		if !ok || !syncFlowToken(key) || seen[key] || key == p["placement_id"] {
			return nil, false
		}
		seen[key] = true
	}
	info := map[string]any{"index_pool": p["index_pool"], "data_extra_pool": p["data_extra_pool"], "index_type": json.Number("0"), "inline_data": true, "storage_classes": map[string]any{"STANDARD": map[string]any{"data_pool": p["data_pool"], "compression_type": p["compression"]}}}
	placements = append(placements, map[string]any{"key": p["placement_id"], "val": info})
	sort.Slice(placements, func(i, j int) bool {
		return syncGroupString(placements[i].(map[string]any), "key") < syncGroupString(placements[j].(map[string]any), "key")
	})
	expected["placement_pools"] = placements
	return expected, true
}
