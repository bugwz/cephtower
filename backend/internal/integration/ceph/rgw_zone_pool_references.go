package ceph

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type rgwPoolReference struct {
	Field     string `json:"field"`
	Pool      string `json:"pool"`
	Namespace string `json:"namespace"`
	Raw       string `json:"raw"`
}

// Match rgw_pool::to_str's canonical escaping, not substring pool-name matches.
// Reject ambiguous input rather than copying from_str's truncation at a second
// unescaped colon. This is configuration evidence, never a pool-delete allowlist.
func parseRGWPoolReference(raw string) (string, string, bool) {
	if raw == "" || len(raw) > 4096 || !utf8.ValidString(raw) || strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return "", "", false
	}
	parts := []string{""}
	escaped := false
	for _, char := range raw {
		if escaped {
			if char != '\\' && char != ':' {
				return "", "", false
			}
			parts[len(parts)-1] += string(char)
			escaped = false
		} else if char == '\\' {
			escaped = true
		} else if char == ':' {
			if len(parts) == 2 {
				return "", "", false
			}
			parts = append(parts, "")
		} else {
			parts[len(parts)-1] += string(char)
		}
	}
	if escaped || parts[0] == "" || (len(parts) == 2 && parts[1] == "") {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}
	return parts[0], parts[1], true
}

func attachRGWZonePoolReferences(zone map[string]any) {
	refs := make([]rgwPoolReference, 0)
	issues := make([]string, 0)
	add := func(field string, value any) {
		raw, ok := value.(string)
		if !ok {
			issues = append(issues, field+": missing or non-string reference")
			return
		}
		if raw == "" {
			return
		} // native empty rgw_pool: not configured
		pool, namespace, ok := parseRGWPoolReference(raw)
		if !ok {
			issues = append(issues, field+": noncanonical pool reference")
			return
		}
		refs = append(refs, rgwPoolReference{Field: field, Pool: pool, Namespace: namespace, Raw: raw})
	}
	known := strings.Fields("domain_root control_pool dedup_pool gc_pool lc_pool log_pool intent_log_pool usage_log_pool roles_pool reshard_pool user_keys_pool user_email_pool user_swift_pool user_uid_pool otp_pool notif_pool topics_pool account_pool group_pool bucket_logging_pool restore_pool")
	seen := map[string]bool{}
	for _, field := range known {
		seen[field] = true
		add(field, zone[field])
	}
	for field, value := range zone {
		if strings.HasSuffix(field, "_pool") && !seen[field] {
			add(field, value)
		}
	}
	placements, ok := zone["placement_pools"].([]any)
	if !ok {
		issues = append(issues, "placement_pools: missing or non-array configuration")
	} else {
		seenPlacement := map[string]bool{}
		for _, value := range placements {
			entry, ok := value.(map[string]any)
			if !ok {
				issues = append(issues, "placement_pools: invalid entry")
				continue
			}
			name, ok := entry["key"].(string)
			if !ok || name == "" || seenPlacement[name] {
				issues = append(issues, "placement_pools: missing or duplicate identity")
				continue
			}
			seenPlacement[name] = true
			config, ok := entry["val"].(map[string]any)
			if !ok {
				issues = append(issues, "placement_pools: invalid placement configuration")
				continue
			}
			prefix := "placement_pools[" + name + "]"
			add(prefix+".index_pool", config["index_pool"])
			add(prefix+".data_extra_pool", config["data_extra_pool"])
			classes, ok := config["storage_classes"].(map[string]any)
			if !ok {
				issues = append(issues, prefix+".storage_classes: missing or non-object configuration")
				continue
			}
			for class, value := range classes {
				config, ok := value.(map[string]any)
				if !ok || class == "" {
					issues = append(issues, prefix+".storage_classes: invalid class")
					continue
				}
				if pool, present := config["data_pool"]; present {
					add(prefix+".storage_classes["+class+"].data_pool", pool)
				}
			}
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Field < refs[j].Field })
	sort.Strings(issues)
	zone["pool_references"] = refs
	zone["pool_reference_issues"] = issues
	zone["pool_references_complete"] = len(issues) == 0
}
