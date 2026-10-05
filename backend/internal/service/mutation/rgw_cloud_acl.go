package mutation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

type cloudACL struct {
	Source      string `json:"source_id"`
	Destination string `json:"dest_id"`
	Type        string `json:"type"`
}

func cloudACLList(value any) ([]cloudACL, bool) {
	raw, err := json.Marshal(value)
	if err != nil || len(raw) == 0 || raw[0] != '[' {
		return nil, false
	}
	var entries []map[string]any
	if json.Unmarshal(raw, &entries) != nil || len(entries) > 256 {
		return nil, false
	}
	result := make([]cloudACL, 0, len(entries))
	seen := map[string]bool{}
	for _, entry := range entries {
		s, sok := entry["source_id"].(string)
		d, dok := entry["dest_id"].(string)
		typ, tok := entry["type"].(string)
		if len(entry) != 3 || !sok || !dok || !tok || s == "" || seen[s] || typ != "id" && typ != "email" && typ != "uri" {
			return nil, false
		}
		for _, v := range []string{s, d} {
			if len(v) > 4096 || strings.ContainsAny(v, ",{}") || strings.IndexFunc(v, unicode.IsControl) >= 0 {
				return nil, false
			}
		}
		seen[s] = true
		result = append(result, cloudACL{s, d, typ})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Source < result[j].Source })
	return result, true
}

func buildCloudACL(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_acl"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	if !syncFlowToken(syncGroupString(p, "realm_id")) || p["storage_class"] == "STANDARD" || p["tier_type"] != "cloud-s3" && p["tier_type"] != "cloud-s3-glacier" {
		return command{}, invalid("existing Realm and cloud tier required")
	}
	old, oldOK := cloudACLList(p["expected_acls"])
	desired, ok := cloudACLList(p["acls"])
	if _, valid := p["confirm_clear"].(bool); !valid {
		return command{}, invalid("explicit clear decision required")
	}
	if !oldOK || !ok || reflect.DeepEqual(old, desired) {
		return command{}, invalid("distinct explicit ACL arrays with unique sources and native types required")
	}
	if len(desired) == 0 && p["confirm_clear"] != true {
		return command{}, invalid("explicit ACL clear confirmation required")
	}
	quoted := func(s string) string { raw, _ := json.Marshal(s); return string(raw) }
	add, remove := []string{}, []string{}
	sources := map[string]bool{}
	for i, a := range desired {
		sources[a.Source] = true
		// JSONFormattable parses values as JSON: quote identities to preserve literal strings.
		add = append(add, fmt.Sprintf("acls[%d].source_id=%s", i, quoted(a.Source)), fmt.Sprintf("acls[%d].dest_id=%s", i, quoted(a.Destination)), fmt.Sprintf("acls[%d].type=%s", i, quoted(a.Type)))
	}
	for _, a := range old {
		if !sources[a.Source] {
			remove = append(remove, fmt.Sprintf("acls[%d].source_id=%s", len(remove), quoted(a.Source)))
		}
	}
	spec.args[2] = "modify"
	if len(add) > 0 {
		spec.args = append(spec.args, "--tier-config", strings.Join(add, ","))
	}
	if len(remove) > 0 {
		spec.args = append(spec.args, "--tier-config-rm", strings.Join(remove, ","))
	}
	if len(strings.Join(spec.args, " ")) > 65536 {
		return command{}, invalid("ACL command exceeds supported size")
	}
	return spec, nil
}

func cloudACLExpected(group, p map[string]any) (map[string]any, bool) {
	expected, ok := zonegroupPlacementDefaultExpected(group, p)
	if !ok {
		return nil, false
	}
	expected["default_placement"] = group["default_placement"]
	if group["default_placement"] == "" {
		expected["default_placement"] = p["placement_id"]
	}
	old, ok := cloudACLList(p["expected_acls"])
	if !ok {
		return nil, false
	}
	desired, ok := cloudACLList(p["acls"])
	if !ok {
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
		mappings, ok := s3["acl_mappings"].([]any)
		if !ok {
			return nil, false
		}
		current := []any{}
		for _, raw := range mappings {
			entry, ok := raw.(map[string]any)
			val, valid := entry["val"].(map[string]any)
			key, keyOK := entry["key"].(string)
			source, sourceOK := val["source_id"].(string)
			if !ok || !valid || !keyOK || !sourceOK || len(entry) != 2 || key != source {
				return nil, false
			}
			current = append(current, val)
		}
		actual, ok := cloudACLList(current)
		if !ok || !reflect.DeepEqual(actual, old) {
			return nil, false
		}
		out := make([]any, 0, len(desired))
		for _, a := range desired {
			out = append(out, map[string]any{"key": a.Source, "val": map[string]any{"source_id": a.Source, "dest_id": a.Destination, "type": a.Type}})
		}
		s3["acl_mappings"] = out
		return expected, true
	}
	return nil, false
}
