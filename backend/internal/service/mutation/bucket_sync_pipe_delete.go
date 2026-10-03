package mutation

func bucketSyncPipeDeleteArgs(p map[string]any) ([]string, error) {
	id := syncGroupString(p, "pipe_id")
	if !syncFlowToken(id) || syncGroupString(p, "expected_group") == "" {
		return nil, invalid("valid pipe_id and expected_group are required")
	}
	// Do not forward selectors: those make native remove modify only part of a pipe.
	return []string{"sync", "group", "pipe", "remove", "--group-id", syncGroupString(p, "group_id"), "--pipe-id", id}, nil
}

func removeBucketSyncPipe(group map[string]any, p map[string]any) error {
	pipes := group["pipes"].([]any)
	remaining := []any{}
	matches := 0
	for _, raw := range pipes {
		pipe, ok := raw.(map[string]any)
		if !ok {
			return invalid("pipe data is invalid")
		}
		id, ok := pipe["id"].(string)
		if !ok {
			return invalid("pipe ID is invalid")
		}
		if id == syncGroupString(p, "pipe_id") {
			matches++
		} else {
			remaining = append(remaining, raw)
		}
	}
	if matches != 1 {
		return invalid("pipe missing or ambiguous; refresh before deleting")
	}
	group["pipes"] = remaining
	return nil
}
