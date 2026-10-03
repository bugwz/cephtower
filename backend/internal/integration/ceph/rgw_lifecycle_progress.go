package ceph

// The native LC key is tenant:name:marker, not a bucket-name substring.
func rgwLifecycleProgress(value any) (map[string]map[string]any, bool) {
	entries, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := map[string]map[string]any{}
	for _, entry := range entries {
		row, ok := entry.(map[string]any)
		if !ok {
			return nil, false
		}
		bucket, bucketOK := row["bucket"].(string)
		status, statusOK := row["status"].(string)
		if !bucketOK || bucket == "" || !statusOK || status == "" {
			return nil, false
		}
		if _, exists := result[bucket]; exists {
			return nil, false
		}
		var started any
		if value, exists := row["started"]; exists {
			text, ok := value.(string)
			if !ok {
				return nil, false
			}
			started = text
		}
		result[bucket] = map[string]any{"bucket": bucket, "status": status, "started": started}
	}
	return result, true
}

func rgwBucketLifecycleProgress(details map[string]any, entries map[string]map[string]any, available bool) any {
	marker, ok := details["marker"].(string)
	if !available || !ok || marker == "" {
		return nil
	}
	key := textField(details, "tenant") + ":" + textField(details, "bucket") + ":" + marker
	entry, found := entries[key]
	if !found {
		return map[string]any{"found": false, "entry": nil}
	}
	return map[string]any{"found": true, "entry": entry}
}
