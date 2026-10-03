package mutation

import (
	"encoding/json"
	"strings"
)

func zonegroupPipeMembership(wanted, group, p map[string]any, spec command) ([]zonegroupSyncWrite, error) {
	var selected map[string]any
	for _, raw := range group["pipes"].([]any) {
		pipe, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid("pipe data invalid")
		}
		id, ok := pipe["id"].(string)
		if !ok {
			return nil, invalid("pipe ID invalid")
		}
		if id == syncGroupString(p, "pipe_id") {
			if selected != nil {
				return nil, invalid("pipe ID ambiguous")
			}
			selected = pipe
		}
	}
	if selected == nil {
		return nil, invalid("pipe missing")
	}
	body, _ := json.Marshal(wanted)
	changes := map[string]pipeZoneChange{}
	for _, side := range []string{"source", "dest"} {
		entity, ok := selected[side].(map[string]any)
		if !ok {
			return nil, invalid("pipe selector unavailable")
		}
		change, err := planSyncPipeZones(entity, p[side+"_zones"], body, true)
		if err != nil {
			return nil, err
		}
		changes[side] = change
	}
	writes := []zonegroupSyncWrite{}
	for _, stage := range []string{"add", "remove"} {
		args := append([]string{}, spec.args...)
		if stage == "remove" {
			args[3] = "remove"
		}
		changed := false
		for _, side := range []string{"source", "dest"} {
			change := changes[side]
			ids, zones := change.added, change.intermediate
			if stage == "remove" {
				ids, zones = change.removed, change.final
			}
			if len(ids) == 0 {
				continue
			}
			changed = true
			args = append(args, "--"+side+"-zone-ids", strings.Join(ids, ","))
			change.entity["zones"] = zones
		}
		if !changed {
			continue
		}
		snapshot, _ := json.Marshal(wanted)
		writes = append(writes, zonegroupSyncWrite{stage: stage, checkStage: stage + "_post_check", args: args, wanted: periodDocument(snapshot)})
	}
	if len(writes) == 0 {
		return nil, invalid("zone membership unchanged")
	}
	return writes, nil
}
