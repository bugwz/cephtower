package ceph

// metadata list bucket exhausts the native pages and reports enumeration errors.
// Unlike bucket list, its global enumeration does not silently ignore errors.
func rgwBucketNames(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok || items == nil {
		return nil, false
	}
	names := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		name, ok := item.(string)
		if !ok || name == "" || seen[name] {
			return nil, false
		}
		seen[name] = true
		names = append(names, name)
	}
	return names, true
}
