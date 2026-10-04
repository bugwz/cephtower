package mutation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func syncPipePriority(p map[string]any) (string, bool, error) {
	value, present := p["priority"]
	if !present {
		return "", false, nil
	}
	switch value.(type) {
	case int, int32, int64, float64, json.Number:
	default:
		return "", true, invalid("priority must be a signed 32-bit integer")
	}
	text := fmt.Sprint(value)
	if number, ok := value.(float64); ok {
		text = strconv.FormatFloat(number, 'f', -1, 64)
	}
	n, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return "", true, invalid("priority must be a signed 32-bit integer")
	}
	return strconv.FormatInt(n, 10), true, nil
}

func syncPipeStorageClass(p map[string]any) (string, bool, error) {
	value, present := p["storage_class"]
	if !present {
		return "", false, nil
	}
	text, ok := value.(string)
	if !ok || (text != "" && !syncFlowToken(text)) {
		return "", true, invalid("invalid destination storage class")
	}
	return text, true, nil
}

func syncPipePrefix(p map[string]any) (string, string, error) {
	raw, present := p["prefix_mode"]
	value, hasValue := p["source_prefix"]
	if !present && !hasValue {
		return "", "", nil
	}
	mode, ok := raw.(string)
	if !ok {
		return "", "", invalid("prefix_mode is required")
	}
	if mode == "remove" && !hasValue {
		return mode, "", nil
	}
	text, ok := value.(string)
	if mode != "set" || !ok || !utf8.ValidString(text) || len(text) > 1024 || strings.IndexFunc(text, unicode.IsControl) >= 0 {
		return "", "", invalid("invalid source prefix change")
	}
	return mode, text, nil
}

// Validate the same explicit selectors and identity as creation, but never send
// zone flags. Advanced parameters change only when explicitly supplied.
func bucketSyncPipeUpdateArgs(p map[string]any) ([]string, error) {
	copy := map[string]any{}
	for k, v := range p {
		copy[k] = v
	}
	copy["source_zones"], copy["dest_zones"] = []any{"*"}, []any{"*"}
	args, err := bucketSyncPipeCreateArgs(copy)
	if err != nil {
		return nil, err
	}
	args[3] = "modify"
	result := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--source-zone-ids" || args[i] == "--dest-zone-ids" {
			i++
			continue
		}
		result = append(result, args[i])
	}
	priority, present, err := syncPipePriority(p)
	if err != nil {
		return nil, err
	}
	if present {
		result = append(result, "--priority", priority)
	}
	storageClass, present, err := syncPipeStorageClass(p)
	if err != nil {
		return nil, err
	}
	if present {
		result = append(result, "--storage-class", storageClass)
	}
	mode, prefix, err := syncPipePrefix(p)
	if err != nil {
		return nil, err
	}
	if mode == "set" {
		result = append(result, "--prefix="+prefix)
	}
	if mode == "remove" {
		result = append(result, "--prefix-rm", "true")
	}
	tagArgs, err := syncPipeTagArgs(p)
	if err != nil {
		return nil, err
	}
	result = append(result, tagArgs...)
	return result, nil
}

func updateBucketSyncPipe(group map[string]any, p map[string]any) error {
	var selected map[string]any
	for _, raw := range group["pipes"].([]any) {
		pipe, ok := raw.(map[string]any)
		if !ok {
			return invalid("pipe data is invalid")
		}
		id, ok := pipe["id"].(string)
		if !ok {
			return invalid("pipe ID is invalid")
		}
		if id == syncGroupString(p, "pipe_id") {
			if selected != nil {
				return invalid("pipe ID is ambiguous")
			}
			selected = pipe
		}
	}
	if selected == nil {
		return invalid("pipe is missing")
	}
	params, ok := selected["params"].(map[string]any)
	if !ok {
		return invalid("pipe params unavailable")
	}
	changed := params["mode"] != syncGroupString(p, "mode")
	priority, present, err := syncPipePriority(p)
	if err != nil {
		return err
	}
	if present {
		changed = changed || !reflect.DeepEqual(params["priority"], json.Number(priority))
		params["priority"] = json.Number(priority)
	}
	storageClass, present, err := syncPipeStorageClass(p)
	if err != nil {
		return err
	}
	if present {
		dest, ok := params["dest"].(map[string]any)
		if !ok {
			return invalid("pipe destination params unavailable")
		}
		changed = changed || !reflect.DeepEqual(dest["storage_class"], storageClass)
		dest["storage_class"] = storageClass
	}
	prefixMode, prefix, err := syncPipePrefix(p)
	if err != nil {
		return err
	}
	if prefixMode != "" {
		source, ok := params["source"].(map[string]any)
		if !ok {
			return invalid("pipe source params unavailable")
		}
		filter, ok := source["filter"].(map[string]any)
		if !ok {
			return invalid("pipe filter unavailable")
		}
		current, exists := filter["prefix"]
		if prefixMode == "remove" {
			changed = changed || exists
			delete(filter, "prefix")
		} else {
			changed = changed || !exists || !reflect.DeepEqual(current, prefix)
			filter["prefix"] = prefix
		}
	}
	tagsChanged, err := updateSyncPipeTags(params, p)
	if err != nil {
		return err
	}
	changed = changed || tagsChanged
	if syncGroupString(p, "mode") == "user" {
		changed = changed || params["user"] != syncGroupString(p, "user")
		params["user"] = syncGroupString(p, "user")
	}
	params["mode"] = syncGroupString(p, "mode")
	// System mode intentionally retains params.user, as native modify does.
	for _, side := range []string{"source", "dest"} {
		entity, ok := selected[side].(map[string]any)
		if !ok {
			return invalid("pipe selector unavailable")
		}
		tenant, name, id, err := pipeBucketFields(p, side)
		if err != nil {
			return err
		}
		if tenant == "*" {
			tenant = ""
		}
		if id == "*" {
			id = ""
		}
		key := name
		if tenant != "" {
			key = tenant + "/" + key
		}
		if id != "" {
			key += ":" + id
		}
		changed = changed || !reflect.DeepEqual(entity["bucket"], key)
		entity["bucket"] = key
	}
	if !changed {
		return invalid("pipe configuration unchanged")
	}
	return nil
}
