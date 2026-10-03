package mutation

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode"
)

func pipeZoneIDs(value any) ([]string, error) {
	entries, ok := value.([]any)
	if !ok || len(entries) == 0 {
		return nil, invalid("nonempty zone ID arrays are required")
	}
	result := []string{}
	seen := map[string]bool{}
	for _, raw := range entries {
		id, ok := raw.(string)
		if !ok || !syncFlowToken(id) || strings.ContainsAny(id, ",;=") || strings.IndexFunc(id, unicode.IsSpace) >= 0 || (strings.Contains(id, "*") && (id != "*" || len(entries) != 1)) || seen[id] {
			return nil, invalid("invalid, duplicate or mixed wildcard zone IDs")
		}
		seen[id] = true
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

func pipeBucketFields(p map[string]any, side string) (string, string, string, error) {
	tenant := syncGroupString(p, side+"_tenant")
	name := syncGroupString(p, side+"_bucket")
	id := syncGroupString(p, side+"_bucket_id")
	if id == "" {
		id = "*"
	}
	for _, part := range []string{tenant, name, id} {
		if part != "" && (!syncFlowToken(part) || strings.ContainsAny(part, "/:\\") || strings.IndexFunc(part, unicode.IsSpace) >= 0 || (strings.Contains(part, "*") && part != "*")) {
			return "", "", "", invalid("invalid bucket selector component")
		}
	}
	if name == "" {
		return "", "", "", invalid("explicit source and destination bucket selectors are required")
	}
	return tenant, name, id, nil
}

func bucketSyncPipeCreateArgs(p map[string]any) ([]string, error) {
	if !syncFlowToken(syncGroupString(p, "pipe_id")) || syncGroupString(p, "expected_group") == "" {
		return nil, invalid("valid pipe_id and expected_group are required")
	}
	mode, uid := syncGroupString(p, "mode"), syncGroupString(p, "user")
	if mode != "system" && mode != "user" {
		return nil, invalid("explicit system or user mode is required")
	}
	if mode == "system" && uid != "" {
		return nil, invalid("system mode must not include a user")
	}
	if mode == "user" {
		parts := strings.Split(uid, "$")
		valid := syncFlowToken(uid) && strings.IndexFunc(uid, unicode.IsSpace) < 0
		switch len(parts) {
		case 1:
			valid = valid && parts[0] != ""
		case 2:
			valid = valid && parts[0] != "" && parts[1] != ""
		case 3:
			valid = valid && parts[1] != "" && parts[2] != ""
		default:
			valid = false
		}
		if !valid {
			return nil, invalid("canonical user UID is required for user mode")
		}
	}
	args := []string{"sync", "group", "pipe", "create", "--group-id", syncGroupString(p, "group_id"), "--pipe-id", syncGroupString(p, "pipe_id")}
	for _, side := range []string{"source", "dest"} {
		zones, err := pipeZoneIDs(p[side+"_zones"])
		if err != nil {
			return nil, err
		}
		tenant, name, id, err := pipeBucketFields(p, side)
		if err != nil {
			return nil, err
		}
		args = append(args, "--"+side+"-zone-ids", strings.Join(zones, ","), "--"+side+"-tenant", tenant, "--"+side+"-bucket", name, "--"+side+"-bucket-id", id)
	}
	args = append(args, "--mode", mode)
	if mode == "user" {
		args = append(args, "--uid", uid)
	}
	return args, nil
}

func addBucketSyncPipe(group map[string]any, p map[string]any, zoneDocument []byte) error {
	pipes := group["pipes"].([]any)
	id := syncGroupString(p, "pipe_id")
	for _, raw := range pipes {
		pipe, ok := raw.(map[string]any)
		if !ok {
			return invalid("existing pipe is invalid")
		}
		if _, ok := pipe["id"].(string); !ok {
			return invalid("existing pipe ID is invalid")
		}
		if pipe["id"] == id {
			return invalid("pipe already exists; creation must not modify it")
		}
	}
	added := map[string]any{"id": id}
	for _, side := range []string{"source", "dest"} {
		ids, _ := pipeZoneIDs(p[side+"_zones"])
		zones := []any{}
		if len(ids) == 1 && ids[0] == "*" {
			zones = append(zones, "*")
		} else {
			rawIDs := []any{}
			for _, id := range ids {
				rawIDs = append(rawIDs, id)
			}
			resolved, err := resolveBucketSyncFlow(map[string]any{"flow_type": "symmetrical", "zones": rawIDs}, zoneDocument)
			if err != nil {
				return err
			}
			zones = resolved["zones"].([]any)
		}
		tenant, name, bucketID, _ := pipeBucketFields(p, side)
		if tenant == "*" {
			tenant = ""
		}
		if bucketID == "*" {
			bucketID = ""
		}
		key := name
		if tenant != "" {
			key = tenant + "/" + key
		}
		if bucketID != "" {
			key += ":" + bucketID
		}
		added[side] = map[string]any{"bucket": key, "zones": zones}
	}
	params := map[string]any{"source": map[string]any{"filter": map[string]any{"tags": []any{}}}, "dest": map[string]any{}, "priority": json.Number("0"), "mode": syncGroupString(p, "mode")}
	if syncGroupString(p, "mode") == "user" {
		params["user"] = syncGroupString(p, "user")
	}
	added["params"] = params
	group["pipes"] = append(pipes, added)
	return nil
}
