package ceph

import "encoding/json"

func deviceLSMData(value any) map[string]any {
	source, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	result := map[string]any{}
	for _, key := range []string{"serialNum", "transport", "mediaType", "health", "rpm", "linkSpeed"} {
		switch v := source[key].(type) {
		case string:
			result[key] = v
		case json.Number:
			if key == "rpm" || key == "linkSpeed" {
				result[key] = v.String()
			}
		}
	}
	if leds, ok := source["ledSupport"].(map[string]any); ok {
		values := map[string]string{}
		for _, key := range []string{"IDENTsupport", "IDENTstatus", "FAILsupport", "FAILstatus"} {
			if v, ok := leds[key].(string); ok {
				values[key] = v
			}
		}
		result["ledSupport"] = values
	}
	return result
}
