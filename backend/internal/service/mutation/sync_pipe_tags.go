package mutation

import (
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type syncPipeTag struct{ key, value string }

func syncPipeTagChanges(p map[string]any) (add, remove []syncPipeTag, present bool, err error) {
	lists := []*[]syncPipeTag{&add, &remove}
	seen := map[syncPipeTag]bool{}
	for i, field := range []string{"tags_add", "tags_remove"} {
		raw, exists := p[field]
		if !exists {
			continue
		}
		present = true
		values, ok := raw.([]any)
		if !ok || values == nil || len(values) > 100 {
			return nil, nil, true, invalid("tag changes require arrays of at most 100 pairs")
		}
		for _, value := range values {
			entry, ok := value.(map[string]any)
			if !ok || len(entry) != 2 {
				return nil, nil, true, invalid("invalid tag pair")
			}
			key, keyOK := entry["key"].(string)
			val, valOK := entry["value"].(string)
			valid := func(s string) bool {
				return utf8.ValidString(s) && len(s) <= 1024 && !strings.Contains(s, ",") && strings.IndexFunc(s, unicode.IsControl) < 0
			}
			tag := syncPipeTag{key, val}
			if !keyOK || !valOK || !valid(key) || !valid(val) || strings.Contains(key, "=") || seen[tag] {
				return nil, nil, true, invalid("invalid, duplicate or conflicting tag pair")
			}
			seen[tag] = true
			*lists[i] = append(*lists[i], tag)
		}
	}
	return
}

func syncPipeTagArgs(p map[string]any) ([]string, error) {
	add, remove, _, err := syncPipeTagChanges(p)
	if err != nil {
		return nil, err
	}
	args := []string{}
	for i, list := range [][]syncPipeTag{remove, add} {
		if len(list) == 0 {
			continue
		}
		parts := make([]string, 0, len(list))
		for _, tag := range list {
			parts = append(parts, tag.key+"="+tag.value)
		}
		flag := []string{"--tags-rm=", "--tags-add="}[i]
		args = append(args, flag+strings.Join(parts, ","))
	}
	return args, nil
}

func updateSyncPipeTags(params map[string]any, p map[string]any) (bool, error) {
	add, remove, present, err := syncPipeTagChanges(p)
	if err != nil || !present {
		return false, err
	}
	source, ok := params["source"].(map[string]any)
	if !ok {
		return false, invalid("pipe source params unavailable")
	}
	filter, ok := source["filter"].(map[string]any)
	if !ok {
		return false, invalid("pipe filter unavailable")
	}
	original, ok := filter["tags"].([]any)
	if !ok || original == nil {
		return false, invalid("pipe tags unavailable")
	}
	tags := map[syncPipeTag]bool{}
	for _, raw := range original {
		item, ok := raw.(map[string]any)
		if !ok || len(item) != 2 {
			return false, invalid("invalid current tag")
		}
		key, keyOK := item["key"].(string)
		val, valOK := item["value"].(string)
		tag := syncPipeTag{key, val}
		if !keyOK || !valOK || tags[tag] {
			return false, invalid("invalid or duplicate current tags")
		}
		tags[tag] = true
	}
	for _, tag := range remove {
		delete(tags, tag)
	}
	for _, tag := range add {
		tags[tag] = true
	}
	sorted := make([]syncPipeTag, 0, len(tags))
	for tag := range tags {
		sorted = append(sorted, tag)
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].key != sorted[j].key {
			return sorted[i].key < sorted[j].key
		}
		return sorted[i].value < sorted[j].value
	})
	result := make([]any, 0, len(sorted))
	for _, tag := range sorted {
		result = append(result, map[string]any{"key": tag.key, "value": tag.value})
	}
	changed := !reflect.DeepEqual(original, result)
	filter["tags"] = result
	return changed, nil
}
