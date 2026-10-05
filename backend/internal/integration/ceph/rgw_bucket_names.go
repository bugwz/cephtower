package ceph

// Unbounded metadata, user and account lists exhaust native pages and return
// a complete top-level key array, not a pagination wrapper.
func rgwMetadataKeys(value any) ([]string, bool) {
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
