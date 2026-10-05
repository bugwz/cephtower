package mutation

import (
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

func cloudConnectionEndpoint(value string) bool {
	if value == "" || len(value) > 4096 || strings.ContainsAny(value, "\\?#") || strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return false
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Opaque != "" || strings.HasSuffix(u.Host, ":") {
		return false
	}
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return false
		}
	}
	return true
}

func buildCloudConnection(p map[string]any) (command, error) {
	params := map[string]any{}
	for k, v := range p {
		params[k] = v
	}
	params["confirm_create"] = p["confirm_connection"]
	spec, err := buildZonegroupStorageClass(params)
	if err != nil {
		return command{}, err
	}
	if !syncFlowToken(syncGroupString(p, "realm_id")) || p["storage_class"] == "STANDARD" || p["tier_type"] != "cloud-s3" && p["tier_type"] != "cloud-s3-glacier" || p["credentials_saved"] != true || !cloudConnectionEndpoint(syncGroupString(p, "endpoint")) {
		return command{}, invalid("existing Realm, cloud tier, explicit safe endpoint and saved credential confirmation required")
	}
	parts := []string{"endpoint=" + tierConfigString(p["endpoint"].(string))}
	for _, key := range []string{"access_key", "secret"} {
		value, ok := p[key].(string)
		if !ok || strings.TrimSpace(value) == "" || len(value) > 4096 || strings.Contains(value, "[REDACTED]") || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return command{}, invalid("explicit nonempty unmasked cloud credentials required")
		}
		parts = append(parts, key+"="+tierConfigString(value))
	}
	spec.args[2] = "modify"
	spec.args = append(spec.args, "--tier-config", strings.Join(parts, ","))
	spec.sensitive = map[int]struct{}{len(spec.args) - 1: {}}
	return spec, nil
}

func cloudConnectionExpected(group, p map[string]any) (map[string]any, bool) {
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
		changed := false
		for _, key := range []string{"endpoint", "access_key", "secret"} {
			old, valid := s3[key].(string)
			if !valid {
				return nil, false
			}
			if old != p[key] {
				changed = true
			}
			s3[key] = p[key]
		}
		return expected, changed
	}
	return nil, false
}
