package mutation

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func syncFlowToken(value string) bool {
	return value != "" && len(value) <= 512 && utf8.ValidString(value) && !strings.HasPrefix(value, "-") && strings.IndexFunc(value, unicode.IsControl) < 0
}

func bucketSyncFlowArgs(p map[string]any) ([]string, error) {
	if syncGroupString(p, "expected_group") == "" {
		return nil, invalid("expected_group is required")
	}
	kind := syncGroupString(p, "flow_type")
	args := []string{"sync", "group", "flow", "create", "--group-id", syncGroupString(p, "group_id"), "--flow-type", kind}
	zoneValid := func(v string) bool {
		return syncFlowToken(v) && !strings.ContainsAny(v, ",;=*") && strings.IndexFunc(v, unicode.IsSpace) < 0
	}
	switch kind {
	case "symmetrical":
		id := syncGroupString(p, "flow_id")
		if !syncFlowToken(id) {
			return nil, invalid("invalid flow_id")
		}
		zones, ok := p["zones"].([]any)
		if !ok || len(zones) == 0 {
			return nil, invalid("zones must be a nonempty array of zone IDs")
		}
		seen := map[string]bool{}
		list := []string{}
		for _, raw := range zones {
			zone, ok := raw.(string)
			if !ok || !zoneValid(zone) || seen[zone] {
				return nil, invalid("invalid or duplicate zone ID")
			}
			seen[zone] = true
			list = append(list, zone)
		}
		if syncGroupString(p, "source_zone") != "" || syncGroupString(p, "dest_zone") != "" {
			return nil, invalid("directional fields are not allowed")
		}
		sort.Strings(list)
		return append(args, "--flow-id", id, "--zone-ids", strings.Join(list, ",")), nil
	case "directional":
		source, dest := syncGroupString(p, "source_zone"), syncGroupString(p, "dest_zone")
		if !zoneValid(source) || !zoneValid(dest) || source == dest {
			return nil, invalid("distinct source and destination zone IDs are required")
		}
		if _, ok := p["zones"]; ok {
			return nil, invalid("zones is not allowed for directional flows")
		}
		if syncGroupString(p, "flow_id") != "" {
			return nil, invalid("directional flows do not have a stored flow ID")
		}
		// Native CLI requires --flow-id, but identifies directional rules by the zone pair.
		return append(args, "--flow-id", "directional", "--source-zone-id", source, "--dest-zone-id", dest), nil
	default:
		return nil, invalid("invalid flow_type")
	}
}

func addBucketSyncFlow(group map[string]any, p map[string]any) error {
	data := group["data_flow"].(map[string]any)
	kind := syncGroupString(p, "flow_type")
	var entries []any
	if raw, exists := data[kind]; exists {
		var ok bool
		entries, ok = raw.([]any)
		if !ok {
			return invalid("existing flow data is invalid")
		}
	}
	id := syncGroupString(p, "flow_id")
	source, dest := syncGroupString(p, "source_zone"), syncGroupString(p, "dest_zone")
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return invalid("existing flow entry is invalid")
		}
		if kind == "symmetrical" {
			if _, ok := entry["id"].(string); !ok {
				return invalid("existing flow ID is invalid")
			}
			if entry["id"] == id {
				return invalid("flow already exists; creation must not extend it")
			}
		} else {
			if _, ok := entry["source_zone"].(string); !ok {
				return invalid("existing source zone is invalid")
			}
			if _, ok := entry["dest_zone"].(string); !ok {
				return invalid("existing destination zone is invalid")
			}
			if entry["source_zone"] == source && entry["dest_zone"] == dest {
				return invalid("directional flow already exists")
			}
		}
	}
	var added map[string]any
	if kind == "symmetrical" {
		zones := append([]any(nil), p["zones"].([]any)...)
		added = map[string]any{"id": id, "zones": zones}
	} else {
		added = map[string]any{"source_zone": source, "dest_zone": dest}
	}
	data[kind] = append(entries, added)
	return nil
}

// Policy output resolves IDs to names, but symmetrical sets retain ID ordering.
func resolveBucketSyncFlow(p map[string]any, body []byte) (map[string]any, error) {
	var document struct {
		Zones []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"zones"`
	}
	if json.Unmarshal(body, &document) != nil || len(document.Zones) == 0 {
		return nil, invalid("zonegroup zone mapping is unavailable")
	}
	names := map[string]string{}
	seenNames := map[string]bool{}
	for _, zone := range document.Zones {
		if zone.ID == "" || zone.Name == "" || names[zone.ID] != "" || seenNames[zone.Name] {
			return nil, invalid("zonegroup zone mapping is ambiguous")
		}
		names[zone.ID] = zone.Name
		seenNames[zone.Name] = true
	}
	result := map[string]any{}
	for k, v := range p {
		result[k] = v
	}
	if syncGroupString(p, "flow_type") == "symmetrical" {
		ids := []string{}
		for _, raw := range p["zones"].([]any) {
			ids = append(ids, raw.(string))
		}
		sort.Strings(ids)
		zones := []any{}
		for _, id := range ids {
			if names[id] == "" {
				return nil, invalid("zone ID is not in the current zonegroup")
			}
			zones = append(zones, names[id])
		}
		result["zones"] = zones
	} else {
		for _, key := range []string{"source_zone", "dest_zone"} {
			name := names[syncGroupString(p, key)]
			if name == "" {
				return nil, invalid("zone ID is not in the current zonegroup")
			}
			result[key] = name
		}
	}
	return result, nil
}
