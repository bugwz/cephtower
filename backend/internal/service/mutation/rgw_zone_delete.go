package mutation

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func buildZoneDelete(p map[string]any) (command, error) {
	realm, ok := p["realm_id"].(string)
	if !ok || realm != "" && !syncFlowToken(realm) || !syncFlowToken(syncGroupString(p, "zone_id")) || !syncFlowToken(syncGroupString(p, "name")) || p["confirm_delete"] != true {
		return command{}, invalid("explicit Zone identity, Realm and deletion confirmation required")
	}
	return command{binary: executor.BinaryRGWAdmin, args: []string{"zone", "delete", "--zone-id", syncGroupString(p, "zone_id")}}, nil
}

// Native deletion scans all groups and may ignore group write failures. Verify the
// entire scanned set, including groups outside the target Realm, after deletion.
func (s *Service) executeZoneDelete(ctx context.Context, access executor.ClusterAccess, req Request, spec command) (cephdomain.ActionResult, error) {
	id, name, realm := syncGroupString(req.Parameters, "zone_id"), syncGroupString(req.Parameters, "name"), syncGroupString(req.Parameters, "realm_id")
	fail := func(stage string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: stage, Message: "Zone deletion could not be verified; inspect local groups, Realm and Period before further action; changes may be partially applied", Retryable: false}
	}
	run := func(stage string, args []string, write bool) (map[string]any, int, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: req.Action + "." + stage, Binary: executor.BinaryRGWAdmin, Args: args, Mutating: write, Timeout: time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		return periodDocument(result.Stdout), result.ExitCode, err == nil && result.ExitCode == 0 && (write || len(result.Stderr) == 0)
	}
	list := func(stage, kind string) (map[string]any, []string, bool) {
		doc, _, ok := run(stage, []string{kind, "list", "--format", "json"}, false)
		names, valid := realmSetupStrings(doc[kind+"s"])
		sort.Strings(names)
		for _, n := range names {
			if !syncFlowToken(n) {
				valid = false
			}
		}
		return doc, names, ok && valid
	}
	before, names, ok := list("zones_before", "zone")
	def, valid := before["default_info"].(string)
	if !ok || !valid || def == id || len(names) < 2 {
		return fail("pre_check_failed")
	}
	remaining := []string{}
	found := false
	for _, n := range names {
		if n == name {
			found = true
		} else {
			remaining = append(remaining, n)
		}
	}
	if !found {
		return fail("pre_check_failed")
	}
	zoneArgs := []string{"zone", "get", "--zone-id", id, "--format", "json"}
	zone, _, ok := run("identity", zoneArgs, false)
	if !ok || zone["id"] != id || zone["name"] != name || zone["realm_id"] != realm {
		return fail("pre_check_failed")
	}
	groupList, groupNames, ok := list("groups_before", "zonegroup")
	if !ok {
		return fail("pre_check_failed")
	}
	groups := make([]map[string]any, len(groupNames))
	expected := make([]map[string]any, len(groupNames))
	seen := map[string]bool{}
	for i, n := range groupNames {
		group, _, ok := run("group_before_"+strconv.Itoa(i), []string{"zonegroup", "get", "--rgw-zonegroup", n, "--format", "json"}, false)
		gid := syncGroupString(group, "id")
		if !ok || !syncFlowToken(gid) || seen[gid] || group["name"] != n {
			return fail("pre_check_failed")
		}
		seen[gid] = true
		next, valid := zoneDeleteExpectedGroup(group, id, realm)
		if !valid {
			return fail("pre_check_failed")
		}
		groups[i], expected[i] = group, next
	}
	current := ""
	if realm != "" {
		r, _, ok := run("realm_before", []string{"realm", "get", "--realm-id", realm, "--format", "json"}, false)
		current = syncGroupString(r, "current_period")
		if !ok || r["id"] != realm || !syncFlowToken(current) {
			return fail("pre_check_failed")
		}
		p, _, ok := run("period_before", []string{"period", "get", "--realm-id", realm, "--period", current, "--format", "json"}, false)
		if !ok || p["id"] != current || p["realm_id"] != realm || !syncFlowToken(syncGroupString(p, "master_zone")) || p["master_zone"] == id {
			return fail("pre_check_failed")
		}
		if !zoneDeleteRealmTopology(expected, realm, p) {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "pre_check_failed", Message: "No Zone deletion submitted: every Realm group must have a valid master member and exactly one master group matching the published Period; resolve pending primary changes separately", Retryable: false}
		}
	}
	checked, _, ok := run("identity_recheck", zoneArgs, false)
	if !ok || !reflect.DeepEqual(zone, checked) {
		return fail("pre_check_failed")
	}
	// Re-read all groups immediately before writing. This is not cross-process CAS.
	for i, g := range groups {
		actual, _, ok := run("group_recheck_"+strconv.Itoa(i), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(g, "id"), "--format", "json"}, false)
		if !ok || !reflect.DeepEqual(actual, g) {
			return fail("pre_check_failed")
		}
	}
	checkedList, checkedNames, ok := list("groups_recheck", "zonegroup")
	if !ok || !reflect.DeepEqual(checkedNames, groupNames) || !reflect.DeepEqual(checkedList, groupList) {
		return fail("pre_check_failed")
	}
	checkedZones, checkedZoneNames, ok := list("zones_recheck", "zone")
	if !ok || checkedZones["default_info"] != def || !reflect.DeepEqual(checkedZoneNames, names) {
		return fail("pre_check_failed")
	}
	if _, _, ok := run("delete", spec.args, true); !ok {
		return fail("command_failed")
	}
	if _, code, _ := run("absence", zoneArgs, false); code != 2 {
		return fail("post_check_failed")
	}
	after, afterNames, ok := list("zones_after", "zone")
	if !ok || after["default_info"] != def || !reflect.DeepEqual(afterNames, remaining) {
		return fail("post_check_failed")
	}
	afterList, afterGroups, ok := list("groups_after", "zonegroup")
	if !ok || !reflect.DeepEqual(afterGroups, groupNames) || !reflect.DeepEqual(afterList, groupList) {
		return fail("post_check_failed")
	}
	for i, g := range expected {
		actual, _, ok := run("group_after_"+strconv.Itoa(i), []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(g, "id"), "--format", "json"}, false)
		if !ok || !reflect.DeepEqual(actual, g) {
			return fail("post_check_failed")
		}
	}
	if realm != "" {
		commitReq := Request{Action: req.Action + ".period", Parameters: map[string]any{"realm_id": realm, "expected_current_period": current}}
		commitSpec := command{args: []string{"period", "update", "--commit", "--realm-id", realm, "--format", "json"}, check: []string{"period", "get", "--realm-id", realm, "--format", "json"}}
		if _, err := s.executePeriodCommit(ctx, access, commitReq, commitSpec); err != nil {
			return fail("post_check_failed")
		}
		published, _, ok := run("published", commitSpec.check, false)
		if !ok || !zoneDeletePublished(published, realm, id, expected) {
			return fail("post_check_failed")
		}
	}
	return cephdomain.ActionResult{Details: map[string]any{"zone_id": id, "zone_absence_verified": true, "group_membership_verified": true, "period_published": realm != "", "pools_deleted": false}}, nil
}

func zoneDeleteExpectedGroup(group map[string]any, id, realm string) (map[string]any, bool) {
	members, ok := group["zones"].([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	found := false
	for _, raw := range members {
		m, ok := raw.(map[string]any)
		z := syncGroupString(m, "id")
		if !ok || !syncFlowToken(z) || seen[z] {
			return nil, false
		}
		seen[z] = true
		found = found || z == id
	}
	if !found {
		return group, true
	}
	master := syncGroupString(group, "master_zone")
	if group["realm_id"] != realm || master == id || !seen[master] || len(members) < 2 {
		return nil, false
	}
	next := map[string]any{}
	for k, v := range group {
		next[k] = v
	}
	kept := []any{}
	for _, raw := range members {
		member := raw.(map[string]any)
		if member["id"] == id {
			continue
		}
		copy := map[string]any{}
		for k, v := range member {
			copy[k] = v
		}
		copy["log_data"] = len(members) > 2
		kept = append(kept, copy)
	}
	next["zones"] = kept
	return next, true
}

func zoneDeletePublished(p map[string]any, realm, id string, expected []map[string]any) bool {
	if p["realm_id"] != realm || !syncFlowToken(syncGroupString(p, "id")) || !syncFlowToken(syncGroupString(p, "master_zone")) || p["master_zone"] == id {
		return false
	}
	pm, _ := p["period_map"].(map[string]any)
	groups, ok := pm["zonegroups"].([]any)
	if !ok || len(groups) == 0 {
		return false
	}
	want := map[string]map[string]any{}
	for _, g := range expected {
		if g["realm_id"] == realm {
			want[syncGroupString(g, "id")] = g
		}
	}
	seen := map[string]bool{}
	for _, raw := range groups {
		g, ok := raw.(map[string]any)
		gid := syncGroupString(g, "id")
		w := want[gid]
		if !ok || w == nil || seen[gid] || g["master_zone"] != w["master_zone"] || !reflect.DeepEqual(g["zones"], w["zones"]) {
			return false
		}
		seen[gid] = true
	}
	masterID := syncGroupString(p, "master_zonegroup")
	return len(seen) == len(want) && seen[masterID] && p["master_zone"] == want[masterID]["master_zone"]
}

// update_period validates every group in the Realm, not just affected groups.
// Do not delete first and discover an invalid pending topology on publication.
// Changing the published primary is a separate operation requiring explicit consent.
func zoneDeleteRealmTopology(groups []map[string]any, realm string, published map[string]any) bool {
	masterID, masterZone := "", ""
	seenGroups := map[string]bool{}
	for _, group := range groups {
		if group["realm_id"] != realm {
			continue
		}
		gid := syncGroupString(group, "id")
		master := syncGroupString(group, "master_zone")
		isMaster, valid := group["is_master"].(bool)
		members, ok := group["zones"].([]any)
		if !valid || !ok || !syncFlowToken(gid) || seenGroups[gid] || !syncFlowToken(master) {
			return false
		}
		seenGroups[gid] = true
		seen := map[string]bool{}
		for _, raw := range members {
			member, ok := raw.(map[string]any)
			id := syncGroupString(member, "id")
			if !ok || !syncFlowToken(id) || seen[id] {
				return false
			}
			seen[id] = true
		}
		if !seen[master] {
			return false
		}
		if isMaster {
			if masterID != "" {
				return false
			}
			masterID, masterZone = gid, master
		}
	}
	return masterID != "" && published["master_zonegroup"] == masterID && published["master_zone"] == masterZone
}
