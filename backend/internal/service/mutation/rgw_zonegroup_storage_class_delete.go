package mutation

func buildZonegroupStorageClassDelete(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_delete"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	spec.args[2] = "rm"
	return spec, nil
}

func zonegroupStorageClassDeleteExpected(group, p map[string]any) (map[string]any, bool) {
	expected, ok := zonegroupPlacementDefaultExpected(group, p)
	if !ok {
		return nil, false
	}
	selectedDefault := expected["default_placement"]
	expected["default_placement"] = group["default_placement"]
	if group["default_placement"] == selectedDefault {
		expected["default_placement"] = p["placement_id"]
	}
	for _, raw := range expected["placement_targets"].([]any) {
		target := raw.(map[string]any)
		if target["name"] != p["placement_id"] {
			continue
		}
		classes, _ := realmSetupStrings(target["storage_classes"])
		remaining := make([]any, 0, len(classes))
		for _, class := range classes {
			if class != p["storage_class"] {
				remaining = append(remaining, class)
			}
		}
		if len(remaining) == 0 {
			// Native binary and JSON decoding restore STANDARD for an empty set.
			if p["storage_class"] == "STANDARD" {
				return nil, false
			}
			remaining = append(remaining, "STANDARD")
		}
		target["storage_classes"] = remaining
		if rawTiers, present := target["tier_targets"]; present {
			tiers, ok := rawTiers.([]any)
			if !ok {
				return nil, false
			}
			kept := make([]any, 0, len(tiers))
			seen := map[string]bool{}
			for _, rawTier := range tiers {
				tier, ok := rawTier.(map[string]any)
				key := syncGroupString(tier, "key")
				if !ok || !syncFlowToken(key) || seen[key] {
					return nil, false
				}
				seen[key] = true
				if key != p["storage_class"] {
					kept = append(kept, tier)
				}
			}
			if len(kept) == 0 {
				delete(target, "tier_targets")
			} else {
				target["tier_targets"] = kept
			}
		}
	}
	return expected, true
}
