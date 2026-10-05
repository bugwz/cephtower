package mutation

import (
	"encoding/json"
)

func buildZonegroupPlacementDefault(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_default"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	spec.args[2] = "default"
	return spec, nil
}

func zonegroupPlacementDefaultExpected(group, p map[string]any) (map[string]any, bool) {
	raw, err := json.Marshal(group)
	if err != nil {
		return nil, false
	}
	expected := periodDocument(raw)
	clear(raw)
	targets, ok := expected["placement_targets"].([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	found := false
	for _, raw := range targets {
		target, ok := raw.(map[string]any)
		name := syncGroupString(target, "name")
		if !ok || !syncFlowToken(name) || seen[name] {
			return nil, false
		}
		seen[name] = true
		if name != p["placement_id"] {
			continue
		}
		classes, ok := realmSetupStrings(target["storage_classes"])
		if !ok {
			return nil, false
		}
		for _, class := range classes {
			found = found || class == p["storage_class"]
		}
	}
	if !found {
		return nil, false
	}
	value := syncGroupString(p, "placement_id")
	if p["storage_class"] != "STANDARD" {
		value += "/" + syncGroupString(p, "storage_class")
	}
	expected["default_placement"] = value
	return expected, true
}
