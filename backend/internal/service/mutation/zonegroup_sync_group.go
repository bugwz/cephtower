package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"time"
)

func zonegroupSyncGroupCommand(action string, p map[string]any, rgw func([]string, []string) command) (command, error) {
	for _, key := range []string{"zonegroup_id", "name", "group_id"} {
		if !syncFlowToken(syncGroupString(p, key)) {
			return command{}, invalid("valid zonegroup identity and group_id are required")
		}
	}
	realm, ok := p["realm_id"].(string)
	if !ok || (realm != "" && !syncFlowToken(realm)) {
		return command{}, invalid("explicit realm_id is required, empty for standalone zonegroup")
	}
	status := syncGroupString(p, "status")
	if action == "rgw_zonegroup.sync_pipe_zones" {
		args, err := bucketSyncPipeZonesArgs(p)
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	if action == "rgw_zonegroup.sync_pipe_update" {
		args, err := bucketSyncPipeUpdateArgs(p)
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	if action == "rgw_zonegroup.sync_pipe_delete" {
		args, err := bucketSyncPipeDeleteArgs(p)
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	if action == "rgw_zonegroup.sync_pipe_create" {
		args, err := bucketSyncPipeCreateArgs(p)
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	if action == "rgw_zonegroup.sync_flow_delete" {
		args, err := bucketSyncFlowDeleteArgs(p)
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	if action == "rgw_zonegroup.sync_flow_update" {
		args, err := bucketSyncFlowArgs(symmetricalFlowParameters(p))
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	if action == "rgw_zonegroup.sync_flow_create" {
		args, err := bucketSyncFlowArgs(p)
		if err != nil {
			return command{}, err
		}
		return rgw(append(args, "--zonegroup-id", syncGroupString(p, "zonegroup_id")), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}), nil
	}
	deleting := action == "rgw_zonegroup.sync_group_delete"
	if !deleting && status != "enabled" && status != "allowed" && status != "forbidden" {
		return command{}, invalid("invalid sync group status")
	}
	verb, expectedKey := "modify", "expected_group"
	if action == "rgw_zonegroup.sync_group_create" {
		verb, expectedKey = "create", "expected_policy"
	}
	if deleting {
		verb = "remove"
	}
	if syncGroupString(p, expectedKey) == "" {
		return command{}, invalid(expectedKey + " is required")
	}
	target := []string{"--zonegroup-id", syncGroupString(p, "zonegroup_id")}
	args := []string{"sync", "group", verb, "--group-id", syncGroupString(p, "group_id")}
	if !deleting {
		args = append(args, "--status", status)
	}
	return rgw(append(args, target...), append([]string{"zonegroup", "get"}, target...)), nil
}

func (s *Service) executeZonegroupSyncGroup(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	run := func(stage string, args []string, write bool) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: args, Mutating: write, Timeout: 2 * time.Minute, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := run("pre_check", spec.check, false)
	wanted := periodDocument(before.Stdout)
	p := request.Parameters
	realm := syncGroupString(p, "realm_id")
	zonegroup := syncGroupString(p, "zonegroup_id")
	if err != nil || wanted == nil || wanted["id"] != zonegroup || wanted["name"] != syncGroupString(p, "name") || wanted["realm_id"] != realm {
		return fail("pre_check_failed", "zonegroup identity or realm changed; no write submitted")
	}
	body, _ := json.Marshal(wanted["sync_policy"])
	currentPolicy, groups, valid := bucketSyncPolicyDocument(body)
	if !valid {
		return fail("pre_check_failed", "sync policy unavailable; no write submitted")
	}
	policy := wanted["sync_policy"].(map[string]any)
	id := syncGroupString(p, "group_id")
	var writes []zonegroupSyncWrite
	if request.Action == "rgw_zonegroup.sync_group_create" {
		expectedPolicy, _, valid := bucketSyncPolicyDocument([]byte(syncGroupString(p, "expected_policy")))
		if !valid || !reflect.DeepEqual(currentPolicy, expectedPolicy) || groups[id] != nil {
			return fail("pre_check_failed", "policy changed or group already exists; refresh before creating")
		}
		entries := append(policy["groups"].([]any), map[string]any{"id": id, "status": syncGroupString(p, "status"), "data_flow": map[string]any{}, "pipes": []any{}})
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].(map[string]any)["id"].(string) < entries[j].(map[string]any)["id"].(string)
		})
		policy["groups"] = entries
	} else {
		_, expected, expectedValid := bucketSyncPolicyDocument([]byte(`{"groups":[` + syncGroupString(p, "expected_group") + `]}`))
		group := groups[id]
		if !valid || !expectedValid || len(expected) != 1 || group == nil || !reflect.DeepEqual(group, expected[id]) {
			return fail("pre_check_failed", "sync group missing or changed; refresh before editing")
		}
		deleting := request.Action == "rgw_zonegroup.sync_group_delete"
		flow := request.Action == "rgw_zonegroup.sync_flow_create"
		flowDelete := request.Action == "rgw_zonegroup.sync_flow_delete"
		flowUpdate := request.Action == "rgw_zonegroup.sync_flow_update"
		pipeCreate := request.Action == "rgw_zonegroup.sync_pipe_create"
		pipeDelete := request.Action == "rgw_zonegroup.sync_pipe_delete"
		pipeUpdate := request.Action == "rgw_zonegroup.sync_pipe_update"
		pipeZones := request.Action == "rgw_zonegroup.sync_pipe_zones"
		if flow {
			// Validate membership, but keep native IDs: zonegroup get is not name-formatted.
			if _, err := resolveBucketSyncFlow(p, before.Stdout); err != nil {
				return fail("pre_check_failed", err.Error())
			}
		}
		if !deleting && !flow && !flowDelete && !flowUpdate && !pipeCreate && !pipeDelete && !pipeUpdate && !pipeZones && group["status"] == syncGroupString(p, "status") {
			return fail("pre_check_failed", "sync group status unchanged")
		}
		// Update the original array, not the canonical index used for comparison.
		remaining := []any{}
		for _, raw := range policy["groups"].([]any) {
			g := raw.(map[string]any)
			if g["id"] == id {
				if pipeZones {
					writes, err = zonegroupPipeMembership(wanted, g, p, spec)
					if err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if pipeUpdate {
					if err := updateBucketSyncPipe(g, p); err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if pipeDelete {
					if err := removeBucketSyncPipe(g, p); err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if pipeCreate {
					if err := addSyncPipe(g, p, before.Stdout, false); err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if flowUpdate {
					writes, err = zonegroupFlowMembership(wanted, g, p, spec)
					if err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if flowDelete {
					// Native zonegroup snapshots contain IDs, including orphaned zones.
					if err := removeBucketSyncFlow(g, p); err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if flow {
					params := map[string]any{}
					for key, value := range p {
						params[key] = value
					}
					if syncGroupString(p, "flow_type") == "symmetrical" {
						zones := append([]any{}, p["zones"].([]any)...)
						sort.Slice(zones, func(i, j int) bool { return zones[i].(string) < zones[j].(string) })
						params["zones"] = zones
					}
					if err := addBucketSyncFlow(g, params); err != nil {
						return fail("pre_check_failed", err.Error())
					}
					remaining = append(remaining, raw)
					continue
				}
				if deleting {
					continue
				}
				g["status"] = syncGroupString(p, "status")
			}
			remaining = append(remaining, raw)
		}
		policy["groups"] = remaining
	}
	current := ""
	if realm != "" {
		result, err := run("realm_pre_check", []string{"realm", "get", "--realm-id", realm, "--format", "json"}, false)
		r := periodDocument(result.Stdout)
		current, _ = r["current_period"].(string)
		if err != nil || r["id"] != realm || !syncFlowToken(current) {
			return fail("pre_check_failed", "realm current period unavailable; no write submitted")
		}
	}
	if len(writes) == 0 {
		writes = []zonegroupSyncWrite{{stage: "write", checkStage: "post_check", args: spec.args, wanted: wanted}}
	}
	for _, write := range writes {
		if _, err := run(write.stage, write.args, true); err != nil {
			return fail("command_failed", "zonegroup "+write.stage+" outcome uncertain; changes may be partial, inspect before any manual retry")
		}
		after, err := run(write.checkStage, spec.check, false)
		if err != nil || !reflect.DeepEqual(periodDocument(after.Stdout), write.wanted) {
			return fail("post_check_failed", "zonegroup "+write.stage+" submitted but full configuration did not match; period not submitted, inspect partial state")
		}
	}
	if realm == "" {
		return cephdomain.ActionResult{}, nil
	}
	commitRequest := Request{Action: request.Action + ".period", Parameters: map[string]any{"realm_id": realm, "expected_current_period": current}}
	commitSpec := command{args: []string{"period", "update", "--commit", "--realm-id", realm, "--format", "json"}, check: []string{"period", "get", "--realm-id", realm, "--format", "json"}}
	if _, err := s.executePeriodCommit(ctx, access, commitRequest, commitSpec); err != nil {
		return fail("post_check_failed", "zonegroup changed but period publication failed or is uncertain; inspect partial state before retrying: "+err.Error())
	}
	published, err := run("published_policy_check", commitSpec.check, false)
	document := periodDocument(published.Stdout)
	if document == nil || !syncFlowToken(syncGroupString(document, "id")) {
		return fail("post_check_failed", "published period identity unavailable; inspect partial state")
	}
	periodMap, _ := document["period_map"].(map[string]any)
	entries, ok := periodMap["zonegroups"].([]any)
	matches := 0
	if err == nil && ok && document["realm_id"] == realm {
		for _, raw := range entries {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if entry["id"] == zonegroup {
				matches++
				if !reflect.DeepEqual(entry["sync_policy"], policy) {
					return fail("post_check_failed", "published zonegroup policy does not match requested configuration")
				}
			}
		}
	}
	if matches != 1 {
		return fail("post_check_failed", "published zonegroup policy unavailable or ambiguous; inspect partial state")
	}
	return cephdomain.ActionResult{}, nil
}
