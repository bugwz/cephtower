package ceph

// Keep the full bucket-local policy, including data_flow and pipes. This is not
// a resolved zonegroup policy or a measurement of replication progress.
func rgwBucketSyncPolicy(value any) (map[string]any, bool) {
	policy, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	groups, ok := policy["groups"].([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	for _, value := range groups {
		group, ok := value.(map[string]any)
		if !ok {
			return nil, false
		}
		id, idOK := group["id"].(string)
		status, statusOK := group["status"].(string)
		_, pipesOK := group["pipes"].([]any)
		_, flowOK := group["data_flow"].(map[string]any)
		if !idOK || !statusOK || status == "" || !pipesOK || !flowOK || seen[id] {
			return nil, false
		}
		seen[id] = true
	}
	return policy, true
}
