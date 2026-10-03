package mutation

import (
	"encoding/json"
	"sort"
	"strings"
)

type zonegroupSyncWrite struct {
	stage      string
	checkStage string
	args       []string
	wanted     map[string]any
}

// Plan against raw zone IDs; each step has its own complete configuration snapshot.
func zonegroupFlowMembership(wanted, group, p map[string]any, spec command) ([]zonegroupSyncWrite, error) {
	params := symmetricalFlowParameters(p)
	encoded, _ := json.Marshal(wanted)
	if _, err := resolveBucketSyncFlow(params, encoded); err != nil {
		return nil, err
	}
	entries, ok := group["data_flow"].(map[string]any)["symmetrical"].([]any)
	if !ok {
		return nil, invalid("symmetrical flows unavailable")
	}
	var flow map[string]any
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid("invalid flow entry")
		}
		if entry["id"] == syncGroupString(p, "flow_id") {
			if flow != nil {
				return nil, invalid("ambiguous flow ID")
			}
			flow = entry
		}
	}
	if flow == nil {
		return nil, invalid("flow missing")
	}
	oldParams := symmetricalFlowParameters(p)
	oldParams["zones"] = flow["zones"]
	if _, err := bucketSyncFlowArgs(oldParams); err != nil {
		return nil, invalid("existing zone ID set invalid")
	}
	old := map[string]bool{}
	desired := map[string]bool{}
	for _, raw := range flow["zones"].([]any) {
		old[raw.(string)] = true
	}
	added, removed := []string{}, []string{}
	for _, raw := range p["zones"].([]any) {
		id := raw.(string)
		desired[id] = true
		if !old[id] {
			added = append(added, id)
		}
	}
	for id := range old {
		if !desired[id] {
			removed = append(removed, id)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) == 0 && len(removed) == 0 {
		return nil, invalid("zone membership unchanged")
	}
	steps := []zonegroupSyncWrite{}
	appendStep := func(stage, verb string, ids []string) {
		ordered := []string{}
		for id := range old {
			ordered = append(ordered, id)
		}
		sort.Strings(ordered)
		zones := []any{}
		for _, id := range ordered {
			zones = append(zones, id)
		}
		flow["zones"] = zones
		args := append([]string{}, spec.args...)
		args[3] = verb
		for i, arg := range args {
			if arg == "--zone-ids" {
				args[i+1] = strings.Join(ids, ",")
				break
			}
		}
		body, _ := json.Marshal(wanted)
		steps = append(steps, zonegroupSyncWrite{stage: stage, checkStage: stage + "_post_check", args: args, wanted: periodDocument(body)})
	}
	if len(added) > 0 {
		for _, id := range added {
			old[id] = true
		}
		appendStep("add", "create", added)
	}
	if len(removed) > 0 {
		for _, id := range removed {
			delete(old, id)
		}
		appendStep("remove", "remove", removed)
	}
	return steps, nil
}
