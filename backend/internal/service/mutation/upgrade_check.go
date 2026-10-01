package mutation

import (
	"bytes"
	"encoding/json"
	"io"
)

func upgradeCheckReport(data []byte) (map[string]any, bool) {
	var report map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&report) != nil || report == nil || decoder.Decode(new(any)) != io.EOF {
		return nil, false
	}
	clean := map[string]any{}
	for _, field := range []string{"target_name", "target_id", "target_version"} {
		value, ok := report[field].(string)
		if !ok || value == "" {
			return nil, false
		}
		clean[field] = value
	}
	if digest, exists := report["target_digest"]; exists {
		if _, ok := digest.(string); !ok {
			return nil, false
		}
		clean["target_digest"] = digest
	}
	for _, field := range []string{"up_to_date", "non_ceph_image_daemons"} {
		entries, ok := report[field].([]any)
		if !ok {
			return nil, false
		}
		for _, entry := range entries {
			if value, ok := entry.(string); !ok || value == "" {
				return nil, false
			}
		}
		clean[field] = entries
	}
	entries, ok := report["needs_update"].(map[string]any)
	if !ok {
		return nil, false
	}
	daemons := map[string]any{}
	for name, entry := range entries {
		fields, ok := entry.(map[string]any)
		if !ok || name == "" {
			return nil, false
		}
		current := map[string]any{}
		for _, field := range []string{"current_name", "current_id", "current_version"} {
			value, exists := fields[field]
			if !exists {
				return nil, false
			}
			if value != nil {
				if _, ok := value.(string); !ok {
					return nil, false
				}
			}
			current[field] = value
		}
		daemons[name] = current
	}
	clean["needs_update"] = daemons
	return clean, true
}
