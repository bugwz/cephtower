package clusterinspect

import "encoding/json"

func validConfigurationMetadata(option map[string]any, name string) bool {
	if option["name"] != name {
		return false
	}
	for _, key := range []string{"type", "level", "desc", "long_desc"} {
		if value, present := option[key]; present {
			if _, ok := value.(string); !ok {
				return false
			}
		}
	}
	if value, present := option["can_update_at_runtime"]; present {
		if _, ok := value.(bool); !ok {
			return false
		}
	}
	for _, key := range []string{"tags", "services", "see_also", "enum_values", "flags"} {
		if value, present := option[key]; present {
			items, ok := value.([]any)
			if !ok {
				return false
			}
			for _, item := range items {
				if _, ok := item.(string); !ok {
					return false
				}
			}
		}
	}
	for _, key := range []string{"default", "daemon_default", "min", "max"} {
		if value, present := option[key]; present {
			switch value.(type) {
			case string, bool, json.Number:
			default:
				return false
			}
		}
	}
	return true
}
