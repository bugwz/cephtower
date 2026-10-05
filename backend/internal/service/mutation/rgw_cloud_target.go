package mutation

import (
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
)

var cloudTargetStrings = []string{"region", "host_style", "target_path", "target_storage_class"}
var cloudTargetNumbers = []string{"multipart_sync_threshold", "multipart_min_part_size"}

func cloudTargetValues(value any) (map[string]any, bool) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}
	result := periodDocument(raw)
	if len(result) != 6 {
		return nil, false
	}
	for _, key := range cloudTargetStrings {
		v, ok := result[key].(string)
		if !ok || len(v) > 4096 || strings.IndexFunc(v, unicode.IsControl) >= 0 {
			return nil, false
		}
	}
	if result["host_style"] != "path" && result["host_style"] != "virtual" {
		return nil, false
	}
	for _, key := range cloudTargetNumbers {
		v, ok := cloudRestoreDays(result[key])
		if !ok {
			return nil, false
		}
		result[key] = json.Number(v)
	}
	return result, true
}

func buildCloudTarget(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_target"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	if !syncFlowToken(syncGroupString(p, "realm_id")) || p["storage_class"] == "STANDARD" || p["tier_type"] != "cloud-s3" && p["tier_type"] != "cloud-s3-glacier" {
		return command{}, invalid("existing Realm and cloud tier required")
	}
	old, oldOK := cloudTargetValues(p["expected_target"])
	desired, ok := cloudTargetValues(p["target"])
	if !oldOK || !ok || reflect.DeepEqual(old, desired) {
		return command{}, invalid("distinct complete target settings with exact unsigned sizes and native host style required")
	}
	parts := []string{}
	for _, key := range cloudTargetStrings {
		parts = append(parts, key+"="+tierConfigString(desired[key].(string)))
	}
	for _, key := range cloudTargetNumbers {
		parts = append(parts, key+"="+desired[key].(json.Number).String())
	}
	spec.args[2] = "modify"
	spec.args = append(spec.args, "--tier-config", strings.Join(parts, ","))
	return spec, nil
}

func cloudTargetExpected(group, p map[string]any) (map[string]any, bool) {
	expected, ok := zonegroupPlacementDefaultExpected(group, p)
	if !ok {
		return nil, false
	}
	expected["default_placement"] = group["default_placement"]
	if group["default_placement"] == "" {
		expected["default_placement"] = p["placement_id"]
	}
	old, oldOK := cloudTargetValues(p["expected_target"])
	desired, ok := cloudTargetValues(p["target"])
	if !oldOK || !ok {
		return nil, false
	}
	for _, raw := range expected["placement_targets"].([]any) {
		target := raw.(map[string]any)
		if target["name"] != p["placement_id"] {
			continue
		}
		tiers, ok := target["tier_targets"].([]any)
		if !ok {
			return nil, false
		}
		seen := map[string]bool{}
		var selected map[string]any
		for _, raw := range tiers {
			tier, ok := raw.(map[string]any)
			key := syncGroupString(tier, "key")
			val, valid := tier["val"].(map[string]any)
			if !ok || !valid || !syncFlowToken(key) || seen[key] {
				return nil, false
			}
			seen[key] = true
			if key == p["storage_class"] {
				selected = val
			}
		}
		if selected == nil || selected["tier_type"] != p["tier_type"] || selected["storage_class"] != p["storage_class"] {
			return nil, false
		}
		s3, ok := selected["s3"].(map[string]any)
		if !ok {
			return nil, false
		}
		current := map[string]any{}
		for key := range old {
			current[key] = s3[key]
		}
		actual, ok := cloudTargetValues(current)
		if !ok || !reflect.DeepEqual(actual, old) {
			return nil, false
		}
		for key, value := range desired {
			s3[key] = value
		}
		return expected, true
	}
	return nil, false
}
