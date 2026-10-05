package mutation

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Build every field explicitly: the native default values are not an API contract.
func cloudCreateConfiguration(p map[string]any) (command, map[string]any, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_connection"], params["confirm_restore"] = p["confirm_create"], p["confirm_create"]
	spec, err := buildCloudConnection(params)
	if err != nil {
		return command{}, nil, err
	}
	restore, err := buildCloudRestore(params)
	if err != nil {
		return command{}, nil, err
	}
	target, ok := cloudTargetValues(p["target"])
	if !ok {
		return command{}, nil, invalid("complete explicit cloud target settings required")
	}
	acls, ok := cloudACLList(p["acls"])
	if !ok {
		return command{}, nil, invalid("explicit valid ACL list required")
	}
	parts := []string{spec.args[len(spec.args)-1], restore.args[len(restore.args)-1]}
	for _, key := range cloudTargetStrings {
		parts = append(parts, key+"="+tierConfigString(target[key].(string)))
	}
	for _, key := range cloudTargetNumbers {
		parts = append(parts, key+"="+target[key].(json.Number).String())
	}
	nativeACLs := make([]any, 0, len(acls))
	for i, a := range acls {
		parts = append(parts, fmt.Sprintf("acls[%d].source_id=%s", i, tierConfigString(a.Source)), fmt.Sprintf("acls[%d].dest_id=%s", i, tierConfigString(a.Destination)), fmt.Sprintf("acls[%d].type=%s", i, tierConfigString(a.Type)))
		nativeACLs = append(nativeACLs, map[string]any{"key": a.Source, "val": map[string]any{"source_id": a.Source, "dest_id": a.Destination, "type": a.Type}})
	}
	spec.args[2] = "add"
	spec.args[len(spec.args)-1] = strings.Join(parts, ",")
	spec.args = append(spec.args, "--tier-type", p["tier_type"].(string))
	if len(strings.Join(spec.args, " ")) > 65536 {
		return command{}, nil, invalid("cloud configuration command exceeds supported size")
	}
	// The connection builder marked the tier-config value before tier-type was appended.
	for _, key := range []string{"endpoint", "access_key", "secret"} {
		target[key] = p[key]
	}
	target["acl_mappings"] = nativeACLs
	days, _ := cloudRestoreDays(p["read_through_restore_days"])
	tier := map[string]any{"tier_type": p["tier_type"], "storage_class": p["storage_class"], "s3": target, "retain_head_object": p["retain_head_object"], "allow_read_through": p["allow_read_through"], "read_through_restore_days": json.Number(days), "restore_storage_class": p["restore_storage_class"]}
	if p["tier_type"] == "cloud-s3-glacier" {
		days, _ = cloudRestoreDays(p["glacier_restore_days"])
		tier["s3-glacier"] = map[string]any{"glacier_restore_days": json.Number(days), "glacier_restore_tier_type": p["glacier_restore_tier_type"]}
	}
	return spec, tier, nil
}

func buildCloudCreate(p map[string]any) (command, error) {
	spec, _, err := cloudCreateConfiguration(p)
	return spec, err
}

func cloudCreateExpected(group, p map[string]any) (map[string]any, bool) {
	expected, ok := zonegroupStorageClassExpected(group, p)
	if !ok {
		return nil, false
	}
	_, tier, err := cloudCreateConfiguration(p)
	if err != nil {
		return nil, false
	}
	for _, raw := range expected["placement_targets"].([]any) {
		target := raw.(map[string]any)
		if target["name"] != p["placement_id"] {
			continue
		}
		classes, _ := realmSetupStrings(target["storage_classes"])
		restoreFound := false
		for _, class := range classes {
			restoreFound = restoreFound || class == p["restore_storage_class"] && class != p["storage_class"]
		}
		if !restoreFound {
			return nil, false
		}
		tiers, ok := target["tier_targets"].([]any)
		if !ok {
			return nil, false
		}
		seen := map[string]bool{}
		for _, raw := range tiers {
			entry, ok := raw.(map[string]any)
			key := syncGroupString(entry, "key")
			val, valid := entry["val"].(map[string]any)
			if !ok || !valid || !syncFlowToken(key) || seen[key] || key == p["restore_storage_class"] || val["storage_class"] != key {
				return nil, false
			}
			seen[key] = true
		}
		tiers = append(tiers, map[string]any{"key": p["storage_class"], "val": tier})
		sort.Slice(tiers, func(i, j int) bool {
			return tiers[i].(map[string]any)["key"].(string) < tiers[j].(map[string]any)["key"].(string)
		})
		target["tier_targets"] = tiers
		return expected, true
	}
	return nil, false
}
