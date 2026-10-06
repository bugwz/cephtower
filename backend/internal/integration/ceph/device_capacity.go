package ceph

import (
	"encoding/json"
	"strconv"
	"strings"
)

func deviceCapacity(records ...map[string]any) *uint64 {
	for _, record := range records {
		for _, key := range []string{"size_bytes", "size"} {
			value, present := record[key]
			if !present {
				continue
			}
			var raw string
			switch typed := value.(type) {
			case json.Number:
				raw = typed.String()
				if whole, fraction, found := strings.Cut(raw, "."); found {
					if fraction == "" || strings.Trim(fraction, "0") != "" {
						return nil
					}
					raw = whole
				}
			case string:
				raw = typed
			default:
				return nil
			}
			parsed, err := strconv.ParseUint(raw, 10, 64)
			if err != nil || strconv.FormatUint(parsed, 10) != raw {
				return nil
			}
			return &parsed
		}
	}
	return nil
}
