package mutation

import (
	"encoding/json"
	"reflect"
	"sort"
)

const dashboardSyncGroup = "dashboard_admin_group"
const dashboardSyncFlow = "dashboard_admin_flow"
const dashboardSyncPipe = "dashboard_admin_pipe"

func replicationPrepareParameters(p map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range p {
		result[key] = value
	}
	result["group_id"] = dashboardSyncGroup
	result["status"] = "allowed"
	return result
}

// Match Dashboard's upper-level permission policy, not a bucket replication rule.
func zonegroupReplicationPreparation(wanted, p map[string]any, spec command) ([]zonegroupSyncWrite, error) {
	if syncGroupString(p, "realm_id") == "" {
		return nil, invalid("replication preparation requires an explicit realm")
	}
	zones, ok := wanted["zones"].([]any)
	if !ok || len(zones) == 0 {
		return nil, invalid("zonegroup members unavailable")
	}
	ids := []any{}
	for _, raw := range zones {
		zone, ok := raw.(map[string]any)
		if !ok {
			return nil, invalid("zone member invalid")
		}
		id, ok := zone["id"].(string)
		if !ok {
			return nil, invalid("zone ID invalid")
		}
		ids = append(ids, id)
	}
	fp := map[string]any{"group_id": dashboardSyncGroup, "expected_group": "prepared", "flow_type": "symmetrical", "flow_id": dashboardSyncFlow, "zones": ids}
	flowArgs, err := bucketSyncFlowArgs(fp)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(wanted)
	if _, err := resolveBucketSyncFlow(fp, body); err != nil {
		return nil, err
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].(string) < ids[j].(string) })
	expectedIDs, err := pipeZoneIDs(p["expected_zones"])
	if err != nil {
		return nil, invalid("expected Zone ID snapshot required")
	}
	actualIDs := []string{}
	for _, id := range ids {
		actualIDs = append(actualIDs, id.(string))
	}
	if !reflect.DeepEqual(expectedIDs, actualIDs) {
		return nil, invalid("zone membership changed; refresh before preparing")
	}
	policy := wanted["sync_policy"].(map[string]any)
	var group map[string]any
	for _, raw := range policy["groups"].([]any) {
		g := raw.(map[string]any)
		if g["id"] == dashboardSyncGroup {
			group = g
		}
	}
	if group == nil {
		return nil, invalid("prepared group unavailable")
	}
	writes := []zonegroupSyncWrite{}
	appendStep := func(stage string, args []string) {
		snapshot, _ := json.Marshal(wanted)
		writes = append(writes, zonegroupSyncWrite{stage: stage, checkStage: stage + "_post_check", args: args, wanted: periodDocument(snapshot)})
	}
	appendStep("group", spec.args)
	if err := addBucketSyncFlow(group, fp); err != nil {
		return nil, err
	}
	target := []string{"--zonegroup-id", syncGroupString(p, "zonegroup_id"), "--format", "json"}
	appendStep("flow", append(flowArgs, target...))
	pp := map[string]any{"group_id": dashboardSyncGroup, "pipe_id": dashboardSyncPipe, "expected_group": "prepared", "source_zones": []any{"*"}, "dest_zones": []any{"*"}, "source_bucket": "*", "dest_bucket": "*", "mode": "system"}
	pipeArgs, err := bucketSyncPipeCreateArgs(pp)
	if err != nil {
		return nil, err
	}
	if err := addSyncPipe(group, pp, body, false); err != nil {
		return nil, err
	}
	appendStep("pipe", append(pipeArgs, target...))
	return writes, nil
}
