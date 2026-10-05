package mutation

import (
	"reflect"
	"sort"
	"strings"
)

func placementTagList(value any) ([]string, bool) {
	tags, ok := realmSetupStrings(value)
	if !ok {
		return nil, false
	}
	for _, tag := range tags {
		if !syncFlowToken(tag) || strings.Contains(tag, ",") {
			return nil, false
		}
	}
	sort.Strings(tags)
	return tags, true
}

func buildZonegroupPlacementTags(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_tags"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	tags, ok := placementTagList(p["tags"])
	old, oldOK := placementTagList(p["expected_tags"])
	if !ok || !oldOK {
		return command{}, invalid("explicit unique comma-free tag arrays required")
	}
	spec.args[2] = "modify"
	if len(tags) > 0 {
		spec.args = append(spec.args, "--tags="+strings.Join(tags, ","))
	} else if len(old) > 0 {
		// An empty --tags is ignored by native modify. Remove the checked old set.
		spec.args = append(spec.args, "--tags-rm="+strings.Join(old, ","))
	}
	return spec, nil
}

func zonegroupPlacementTagsExpected(group, p map[string]any) (map[string]any, bool) {
	// Reuse full-document cloning and declared target/class validation.
	expected, ok := zonegroupPlacementDefaultExpected(group, p)
	if !ok {
		return nil, false
	}
	expected["default_placement"] = group["default_placement"]
	if group["default_placement"] == "" {
		expected["default_placement"] = p["placement_id"]
	}
	tags, ok := placementTagList(p["tags"])
	old, oldOK := placementTagList(p["expected_tags"])
	if !ok || !oldOK {
		return nil, false
	}
	for _, raw := range expected["placement_targets"].([]any) {
		target := raw.(map[string]any)
		if target["name"] != p["placement_id"] {
			continue
		}
		current, ok := placementTagList(target["tags"])
		if !ok || !reflect.DeepEqual(current, old) {
			return nil, false
		}
		values := make([]any, len(tags))
		for i, tag := range tags {
			values[i] = tag
		}
		target["tags"] = values
	}
	return expected, true
}
