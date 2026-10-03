package mutation

import (
	"encoding/json"
	"reflect"
	"strings"
)

func rgwUserPlacementRequested(params map[string]any) bool {
	_, placement := params["default_placement"]
	_, tags := params["placement_tags_csv"]
	return placement || tags
}

func rgwUserPlacementMatches(raw []byte, params map[string]any) bool {
	var info map[string]json.RawMessage
	if json.Unmarshal(raw, &info) != nil || info == nil {
		return false
	}
	if _, present := params["default_placement"]; present {
		for _, field := range []string{"default_placement", "default_storage_class"} {
			var actual *string
			expected, ok := params[field].(string)
			if !ok || json.Unmarshal(info[field], &actual) != nil || actual == nil || *actual != expected {
				return false
			}
		}
	}
	if value, present := params["placement_tags_csv"]; present {
		expected, ok := value.(string)
		var actual []string
		if !ok || json.Unmarshal(info["placement_tags"], &actual) != nil || !reflect.DeepEqual(actual, strings.Split(expected, ",")) {
			return false
		}
	}
	return true
}
