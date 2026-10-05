package mutation

import (
	"encoding/json"
	"strconv"
	"strings"
)

func cloudRestoreDays(value any) (string, bool) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", false
	}
	n, err := strconv.ParseUint(string(raw), 10, 64)
	return strconv.FormatUint(n, 10), err == nil && n <= 9007199254740991
}

func buildCloudRestore(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_restore"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	retain, retainOK := p["retain_head_object"].(bool)
	read, readOK := p["allow_read_through"].(bool)
	days, daysOK := cloudRestoreDays(p["read_through_restore_days"])
	restore := syncGroupString(p, "restore_storage_class")
	if !retainOK || !readOK || !daysOK || !syncFlowToken(restore) || strings.ContainsAny(restore, ",= /\\") || !syncFlowToken(syncGroupString(p, "realm_id")) || p["storage_class"] == "STANDARD" {
		return command{}, invalid("existing Realm, cloud class, explicit booleans, exact unsigned days and local restore class required")
	}
	if p["tier_type"] != "cloud-s3" && p["tier_type"] != "cloud-s3-glacier" {
		return command{}, invalid("supported cloud tier required")
	}
	spec.args[2] = "modify"
	spec.args = append(spec.args, "--tier-config", "retain_head_object="+strconv.FormatBool(retain)+",allow_read_through="+strconv.FormatBool(read)+",read_through_restore_days="+days+",restore_storage_class="+restore)
	if p["tier_type"] == "cloud-s3-glacier" {
		glacierDays, valid := cloudRestoreDays(p["glacier_restore_days"])
		level := syncGroupString(p, "glacier_restore_tier_type")
		if !valid || level != "Standard" && level != "Expedited" {
			return command{}, invalid("explicit exact unsigned Glacier days and Standard or Expedited restore tier required")
		}
		spec.args[len(spec.args)-1] += ",glacier_restore_days=" + glacierDays + ",glacier_restore_tier_type=" + level
	} else {
		_, daysPresent := p["glacier_restore_days"]
		_, levelPresent := p["glacier_restore_tier_type"]
		if daysPresent || levelPresent {
			return command{}, invalid("Glacier settings require a Glacier tier")
		}
	}
	return spec, nil
}

func cloudRestoreExpected(group, p map[string]any) (map[string]any, bool) {
	expected, ok := zonegroupPlacementDefaultExpected(group, p)
	if !ok {
		return nil, false
	}
	expected["default_placement"] = group["default_placement"]
	if group["default_placement"] == "" {
		expected["default_placement"] = p["placement_id"]
	}
	for _, raw := range expected["placement_targets"].([]any) {
		target := raw.(map[string]any)
		if target["name"] != p["placement_id"] {
			continue
		}
		classes, ok := realmSetupStrings(target["storage_classes"])
		if !ok {
			return nil, false
		}
		restoreFound := false
		for _, c := range classes {
			restoreFound = restoreFound || c == p["restore_storage_class"]
		}
		if !restoreFound {
			return nil, false
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
			if !ok || !valid || !syncFlowToken(key) || seen[key] || key == p["restore_storage_class"] {
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
		days, ok := cloudRestoreDays(p["read_through_restore_days"])
		if !ok {
			return nil, false
		}
		for _, key := range []string{"retain_head_object", "allow_read_through", "restore_storage_class"} {
			selected[key] = p[key]
		}
		selected["read_through_restore_days"] = json.Number(days)
		if p["tier_type"] == "cloud-s3-glacier" {
			glacier, valid := selected["s3-glacier"].(map[string]any)
			glacierDays, daysOK := cloudRestoreDays(p["glacier_restore_days"])
			if !valid || !daysOK {
				return nil, false
			}
			glacier["glacier_restore_days"] = json.Number(glacierDays)
			glacier["glacier_restore_tier_type"] = p["glacier_restore_tier_type"]
		}
		return expected, true
	}
	return nil, false
}
