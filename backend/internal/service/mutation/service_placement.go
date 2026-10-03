package mutation

import (
	"encoding/json"
	"strings"
)

func validateServicePlacement(p map[string]any) error {
	for key := range p {
		switch key {
		case "count", "count_per_host", "host_pattern", "hosts", "label":
		default:
			return invalid("unknown placement field: " + key)
		}
	}
	for _, key := range []string{"count", "count_per_host"} {
		if value, exists := p[key]; exists {
			raw, err := json.Marshal(value)
			var number *int64
			if err != nil || json.Unmarshal(raw, &number) != nil || number == nil || *number < 1 {
				return invalid(key + " must be a positive integer")
			}
		}
	}
	var label, pattern string
	for key, target := range map[string]*string{"label": &label, "host_pattern": &pattern} {
		if value, exists := p[key]; exists {
			text, ok := value.(string)
			if !ok {
				return invalid(key + " must be a string")
			}
			*target = text
		}
	}
	var hosts []string
	if value, exists := p["hosts"]; exists {
		raw, err := json.Marshal(value)
		if err != nil || json.Unmarshal(raw, &hosts) != nil || hosts == nil {
			return invalid("hosts must be an array of strings")
		}
		for _, host := range hosts {
			if strings.TrimSpace(host) == "" {
				return invalid("host placement cannot be empty")
			}
		}
	}
	if len(hosts) > 0 && (label != "" || pattern != "") {
		return invalid("explicit hosts cannot be combined with label or host_pattern")
	}
	if _, perHost := p["count_per_host"]; perHost {
		if _, count := p["count"]; count {
			return invalid("count and count_per_host are mutually exclusive")
		}
		if label == "" && pattern == "" && len(hosts) == 0 {
			return invalid("count_per_host requires a host selector")
		}
		for _, host := range hosts {
			if strings.ContainsAny(host, ":=") {
				return invalid("count_per_host cannot use named or network-specific host placements")
			}
		}
	}
	return nil
}
