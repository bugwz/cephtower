package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func bucketSyncPipeZonesArgs(p map[string]any) ([]string, error) {
	args, err := bucketSyncPipeDeleteArgs(p)
	if err != nil {
		return nil, err
	}
	for _, side := range []string{"source", "dest"} {
		if _, err := pipeZoneIDs(p[side+"_zones"]); err != nil {
			return nil, err
		}
	}
	args[3] = "modify"
	return args, nil
}

type pipeZoneChange struct {
	entity              map[string]any
	added, removed      []string
	intermediate, final []any
}

// Native add_zones switches wildcard to a concrete set directly. Removing '*'
// afterwards is unnecessary and risks discarding the intended wildcard state.
func planPipeZones(entity map[string]any, desired any, zoneBody []byte) (pipeZoneChange, error) {
	result := pipeZoneChange{entity: entity}
	ids, err := pipeZoneIDs(desired)
	if err != nil {
		return result, err
	}
	// Validate the entire mapping even when both sides use wildcards.
	_, err = resolveBucketSyncFlow(map[string]any{"flow_type": "symmetrical", "zones": []any{}}, zoneBody)
	if err != nil {
		return result, err
	}
	var document struct {
		Zones []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"zones"`
	}
	if json.Unmarshal(zoneBody, &document) != nil {
		return result, invalid("zone mapping invalid")
	}
	byName, byID := map[string]string{}, map[string]string{}
	for _, zone := range document.Zones {
		if zone.Name == "*" {
			return result, invalid("zone name conflicts with wildcard")
		}
		byName[zone.Name] = zone.ID
		byID[zone.ID] = zone.Name
	}
	raw, exists := entity["zones"]
	old := []any{}
	if exists {
		var ok bool
		old, ok = raw.([]any)
		if !ok {
			return result, invalid("existing zones unavailable")
		}
	}
	oldWildcard := len(old) == 1 && old[0] == "*"
	oldIDs := map[string]bool{}
	if !oldWildcard {
		for _, raw := range old {
			name, ok := raw.(string)
			id := byName[name]
			if !ok || id == "" || id == "*" || oldIDs[id] {
				return result, invalid("existing zone mapping unavailable or ambiguous")
			}
			if _, err := pipeZoneIDs([]any{id}); err != nil {
				return result, err
			}
			oldIDs[id] = true
		}
	}
	wildcard := len(ids) == 1 && ids[0] == "*"
	if wildcard {
		result.final = []any{"*"}
		result.intermediate = result.final
		if !oldWildcard {
			result.added = []string{"*"}
		}
		return result, nil
	}
	target := map[string]bool{}
	for _, id := range ids {
		if byID[id] == "" {
			return result, invalid("zone ID is not in current zonegroup")
		}
		target[id] = true
		if !oldIDs[id] {
			result.added = append(result.added, id)
		}
	}
	for id := range oldIDs {
		if !target[id] {
			result.removed = append(result.removed, id)
		}
	}
	sort.Strings(result.removed)
	names := func(set map[string]bool) []any {
		keys := []string{}
		for id := range set {
			keys = append(keys, id)
		}
		sort.Strings(keys)
		out := []any{}
		for _, id := range keys {
			out = append(out, byID[id])
		}
		return out
	}
	result.final = names(target)
	for _, id := range result.added {
		oldIDs[id] = true
	}
	result.intermediate = names(oldIDs)
	return result, nil
}

func (s *Service) executeBucketSyncPipeZones(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	read := func(stage string) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: spec.check, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := read("pre_check")
	if err != nil {
		return fail("pre_check_failed", "policy unavailable; no change submitted")
	}
	wanted, groups, valid := bucketSyncPolicyDocument(before.Stdout)
	_, expected, expectedValid := bucketSyncPolicyDocument([]byte(`{"groups":[` + syncGroupString(request.Parameters, "expected_group") + `]}`))
	group := groups[syncGroupString(request.Parameters, "group_id")]
	if !valid || !expectedValid || len(expected) != 1 || group == nil || !reflect.DeepEqual(group, expected[syncGroupString(request.Parameters, "group_id")]) {
		return fail("pre_check_failed", "group missing or changed; refresh before editing")
	}
	var selected map[string]any
	for _, raw := range group["pipes"].([]any) {
		pipe, ok := raw.(map[string]any)
		if !ok {
			return fail("pre_check_failed", "pipe data invalid")
		}
		id, ok := pipe["id"].(string)
		if !ok {
			return fail("pre_check_failed", "pipe ID invalid")
		}
		if id == syncGroupString(request.Parameters, "pipe_id") {
			if selected != nil {
				return fail("pre_check_failed", "pipe ID ambiguous")
			}
			selected = pipe
		}
	}
	if selected == nil {
		return fail("pre_check_failed", "pipe missing")
	}
	zones, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".zones", Binary: spec.binary, Args: []string{"zonegroup", "get", "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return fail("pre_check_failed", "zonegroup unavailable; no change submitted")
	}
	changes := map[string]pipeZoneChange{}
	for _, side := range []string{"source", "dest"} {
		entity, ok := selected[side].(map[string]any)
		if !ok {
			return fail("pre_check_failed", "pipe selector unavailable")
		}
		change, err := planPipeZones(entity, request.Parameters[side+"_zones"], zones.Stdout)
		if err != nil {
			return fail("pre_check_failed", err.Error())
		}
		changes[side] = change
	}
	wrote := false
	for _, stage := range []string{"add", "remove"} {
		args := append([]string(nil), spec.args...)
		if stage == "remove" {
			args[3] = "remove"
		}
		hasChange := false
		for _, side := range []string{"source", "dest"} {
			change := changes[side]
			ids := change.added
			names := change.intermediate
			if stage == "remove" {
				ids = change.removed
				names = change.final
			}
			if len(ids) == 0 {
				continue
			}
			hasChange = true
			args = append(args, "--"+side+"-zone-ids", strings.Join(ids, ","))
			change.entity["zones"] = names
		}
		if !hasChange {
			continue
		}
		wrote = true
		_, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: spec.binary, Args: args, Mutating: true, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput})
		if err != nil {
			return fail("command_failed", stage+" outcome uncertain; changes may be partial, refresh before retrying")
		}
		after, err := read(stage + "_post_check")
		actual, _, ok := bucketSyncPolicyDocument(after.Stdout)
		if err != nil || !ok || !reflect.DeepEqual(wanted, actual) {
			return fail("post_check_failed", stage+" submitted but policy did not match; stop and inspect partial changes")
		}
	}
	if !wrote {
		return fail("pre_check_failed", "zone membership unchanged")
	}
	return cephdomain.ActionResult{}, nil
}
