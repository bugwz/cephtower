package mutation

import (
	"encoding/json"
	"sort"
	"strings"
)

func buildZonegroupPlacementCreate(p map[string]any) (command, error) {
	params := map[string]any{}
	for key, value := range p {
		params[key] = value
	}
	params["storage_class"] = "STANDARD"
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	tags, ok := realmSetupStrings(p["tags"])
	if !ok {
		return command{}, invalid("explicit unique tag list required")
	}
	for _, tag := range tags {
		if !syncFlowToken(tag) || strings.Contains(tag, ",") {
			return command{}, invalid("placement tags cannot contain commas or control characters")
		}
	}
	if len(tags) > 0 {
		spec.args = append(spec.args, "--tags="+strings.Join(tags, ","))
	}
	return spec, nil
}

func zonegroupPlacementCreateExpected(group, p map[string]any) (map[string]any, bool) {
	raw, err := json.Marshal(group)
	if err != nil {
		return nil, false
	}
	expected := periodDocument(raw)
	clear(raw)
	def, ok := expected["default_placement"].(string)
	if !ok {
		return nil, false
	}
	targets, ok := expected["placement_targets"].([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	for _, raw := range targets {
		target, ok := raw.(map[string]any)
		name := syncGroupString(target, "name")
		if !ok || !syncFlowToken(name) || seen[name] || name == p["placement_id"] {
			return nil, false
		}
		seen[name] = true
	}
	tags, ok := realmSetupStrings(p["tags"])
	if !ok {
		return nil, false
	}
	sort.Strings(tags)
	values := make([]any, len(tags))
	for i, tag := range tags {
		values[i] = tag
	}
	targets = append(targets, map[string]any{"name": p["placement_id"], "tags": values, "storage_classes": []any{"STANDARD"}})
	sort.Slice(targets, func(i, j int) bool {
		return syncGroupString(targets[i].(map[string]any), "name") < syncGroupString(targets[j].(map[string]any), "name")
	})
	expected["placement_targets"] = targets
	if def == "" {
		expected["default_placement"] = p["placement_id"]
	}
	return expected, true
}
