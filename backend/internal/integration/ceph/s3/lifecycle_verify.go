package s3

import (
	"encoding/json"
	"sort"
)

// BucketLifecycleMatches checks all rules, allowing native-generated IDs only
// where the submitted ID was empty. Explicit identities must match first.
func BucketLifecycleMatches(wanted, actual []byte) (bool, error) {
	want, err := BucketLifecycle(wanted)
	if err != nil {
		return false, err
	}
	got, err := BucketLifecycle(actual)
	if err != nil {
		return false, err
	}
	if len(want) != len(got) {
		return false, nil
	}
	byID := map[string]string{}
	for _, rule := range got {
		if rule.ID == "" {
			return false, nil
		}
		if _, exists := byID[rule.ID]; exists {
			return false, nil
		}
		byID[rule.ID] = lifecycleRuleSignature(rule)
	}
	anonymous := map[string]int{}
	for _, rule := range want {
		signature := lifecycleRuleSignature(rule)
		if rule.ID == "" {
			anonymous[signature]++
			continue
		}
		if byID[rule.ID] != signature {
			return false, nil
		}
		delete(byID, rule.ID)
	}
	for _, signature := range byID {
		if anonymous[signature] == 0 {
			return false, nil
		}
		anonymous[signature]--
	}
	return true, nil
}

func lifecycleRuleSignature(rule BucketLifecycleRule) string {
	rule.ID = ""
	// RGW emits And according to condition count and falls back to Prefix for
	// an empty filter. Empty scalar conditions are not emitted.
	rule.Selector.Kind = ""
	rule.Selector.And = false
	for _, field := range []**string{&rule.Selector.Prefix, &rule.Selector.ObjectSizeGreaterThan, &rule.Selector.ObjectSizeLessThan} {
		if *field != nil && **field == "" {
			*field = nil
		}
	}
	sort.Slice(rule.Selector.Tags, func(i, j int) bool {
		left, right := rule.Selector.Tags[i], rule.Selector.Tags[j]
		if left.Key != right.Key {
			return left.Key < right.Key
		}
		return left.Value < right.Value
	})
	actions := make([]string, 0, len(rule.Actions))
	for _, action := range rule.Actions {
		// A false deletion marker carries no operation and is omitted on GET.
		if action.Type == "Expiration" && action.Fields["ExpiredObjectDeleteMarker"] == "false" {
			continue
		}
		encoded, _ := json.Marshal(action)
		actions = append(actions, string(encoded))
	}
	sort.Strings(actions)
	encoded, _ := json.Marshal(struct {
		Status   string
		Selector BucketLifecycleSelector
		Actions  []string
	}{rule.Status, rule.Selector, actions})
	return string(encoded)
}
