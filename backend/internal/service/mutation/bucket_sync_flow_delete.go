package mutation

func bucketSyncFlowDeleteArgs(p map[string]any) ([]string, error) {
	if syncGroupString(p, "flow_type") == "directional" {
		args, err := bucketSyncFlowArgs(p)
		if err != nil {
			return nil, err
		}
		args[3] = "remove"
		return args, nil
	}
	if syncGroupString(p, "flow_type") != "symmetrical" || !syncFlowToken(syncGroupString(p, "flow_id")) || syncGroupString(p, "expected_group") == "" {
		return nil, invalid("valid flow_type, flow_id and expected_group are required")
	}
	if _, exists := p["zones"]; exists {
		return nil, invalid("whole flow deletion does not accept zones")
	}
	if syncGroupString(p, "source_zone") != "" || syncGroupString(p, "dest_zone") != "" {
		return nil, invalid("directional fields are not allowed")
	}
	return []string{"sync", "group", "flow", "remove", "--group-id", syncGroupString(p, "group_id"), "--flow-type", "symmetrical", "--flow-id", syncGroupString(p, "flow_id")}, nil
}

func removeBucketSyncFlow(group map[string]any, p map[string]any) error {
	data := group["data_flow"].(map[string]any)
	kind := syncGroupString(p, "flow_type")
	entries, ok := data[kind].([]any)
	if !ok {
		return invalid("flow list missing or invalid")
	}
	remaining := []any{}
	matches := 0
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return invalid("flow entry invalid")
		}
		match := false
		if kind == "symmetrical" {
			if _, ok := entry["id"].(string); !ok {
				return invalid("flow ID invalid")
			}
			match = entry["id"] == syncGroupString(p, "flow_id")
		} else {
			if _, ok := entry["source_zone"].(string); !ok {
				return invalid("source zone invalid")
			}
			if _, ok := entry["dest_zone"].(string); !ok {
				return invalid("destination zone invalid")
			}
			match = entry["source_zone"] == syncGroupString(p, "source_zone") && entry["dest_zone"] == syncGroupString(p, "dest_zone")
		}
		if match {
			matches++
		} else {
			remaining = append(remaining, raw)
		}
	}
	if matches != 1 {
		return invalid("flow missing or ambiguous; refresh before deleting")
	}
	if len(remaining) == 0 {
		delete(data, kind)
	} else {
		data[kind] = remaining
	}
	return nil
}
