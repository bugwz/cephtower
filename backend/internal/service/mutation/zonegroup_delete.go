package mutation

import (
	"context"
	"reflect"
	"sort"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func buildZonegroupDelete(p map[string]any) (command, error) {
	realm, ok := p["realm_id"].(string)
	zones, valid := realmSetupStrings(p["expected_zones"])
	if !ok || realm != "" && !syncFlowToken(realm) || !valid || p["confirm_delete"] != true || !syncFlowToken(syncGroupString(p, "zonegroup_id")) || !syncFlowToken(syncGroupString(p, "name")) {
		return command{}, invalid("explicit zonegroup identity, realm, member IDs and deletion confirmation required")
	}
	for _, zone := range zones {
		if !syncFlowToken(zone) {
			return command{}, invalid("invalid member Zone ID")
		}
	}
	return command{binary: executor.BinaryRGWAdmin, args: []string{"zonegroup", "delete", "--zonegroup-id", syncGroupString(p, "zonegroup_id")}}, nil
}

func (s *Service) executeZonegroupDelete(ctx context.Context, access executor.ClusterAccess, req Request, spec command) (cephdomain.ActionResult, error) {
	id, name, realm := syncGroupString(req.Parameters, "zonegroup_id"), syncGroupString(req.Parameters, "name"), syncGroupString(req.Parameters, "realm_id")
	fail := func(code, message string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: false}
	}
	run := func(stage string, args []string, write bool) (map[string]any, int, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: req.Action + "." + stage, Binary: executor.BinaryRGWAdmin, Args: args, Mutating: write, Timeout: time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		return periodDocument(result.Stdout), result.ExitCode, err == nil && result.ExitCode == 0 && len(result.Stderr) == 0
	}
	list := func(stage string) ([]string, string, bool) {
		doc, _, ok := run(stage, []string{"zonegroup", "list", "--format", "json"}, false)
		names, valid := realmSetupStrings(doc["zonegroups"])
		def, defOK := doc["default_info"].(string)
		sort.Strings(names)
		return names, def, ok && valid && defOK
	}
	names, def, ok := list("list_before")
	if !ok || def == id || len(names) < 2 {
		return fail("pre_check_failed", "Cannot delete the only zonegroup or current-context default; verify configuration first")
	}
	remaining := []string{}
	found := false
	for _, item := range names {
		if item == name {
			found = true
		} else {
			remaining = append(remaining, item)
		}
	}
	if !found {
		return fail("pre_check_failed", "Zonegroup name not found; no deletion submitted")
	}
	group, _, ok := run("identity", []string{"zonegroup", "get", "--zonegroup-id", id, "--format", "json"}, false)
	if !ok || group["id"] != id || group["name"] != name || group["realm_id"] != realm || group["is_master"] != false {
		return fail("pre_check_failed", "Zonegroup identity, Realm or master status changed; no deletion submitted")
	}
	members, ok := group["zones"].([]any)
	actual := []string{}
	seen := map[string]bool{}
	if !ok {
		return fail("pre_check_failed", "Zonegroup membership unavailable")
	}
	for _, raw := range members {
		member, ok := raw.(map[string]any)
		zone := syncGroupString(member, "id")
		if !ok || !syncFlowToken(zone) || seen[zone] {
			return fail("pre_check_failed", "Zonegroup membership invalid")
		}
		seen[zone] = true
		actual = append(actual, zone)
	}
	expected, _ := realmSetupStrings(req.Parameters["expected_zones"])
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		return fail("pre_check_failed", "Zonegroup membership changed; refresh before deletion")
	}
	current := ""
	if realm != "" {
		r, _, ok := run("realm_before", []string{"realm", "get", "--realm-id", realm, "--format", "json"}, false)
		current = syncGroupString(r, "current_period")
		if !ok || r["id"] != realm || !syncFlowToken(current) {
			return fail("pre_check_failed", "Realm current Period unavailable")
		}
		period, _, ok := run("period_before", []string{"period", "get", "--realm-id", realm, "--period", current, "--format", "json"}, false)
		master := syncGroupString(period, "master_zonegroup")
		if !ok || period["realm_id"] != realm || period["id"] != current || !syncFlowToken(master) || master == id {
			return fail("pre_check_failed", "Published master zonegroup is unavailable or is the deletion target")
		}
		masterGroup, _, ok := run("master_before", []string{"zonegroup", "get", "--zonegroup-id", master, "--format", "json"}, false)
		if !ok || masterGroup["id"] != master || masterGroup["realm_id"] != realm || masterGroup["is_master"] != true {
			return fail("pre_check_failed", "Remaining Realm master zonegroup is not verifiable")
		}
	}
	// Re-read after the preflight sequence; not a cross-process CAS.
	checked, _, ok := run("identity_recheck", []string{"zonegroup", "get", "--zonegroup-id", id, "--format", "json"}, false)
	if !ok || !reflect.DeepEqual(checked, group) {
		return fail("pre_check_failed", "Zonegroup changed during preflight")
	}
	if _, _, ok := run("delete", spec.args, true); !ok {
		return fail("command_failed", "Zonegroup deletion failed or partially applied; inspect before any further action")
	}
	_, code, _ := run("absence", []string{"zonegroup", "get", "--zonegroup-id", id, "--format", "json"}, false)
	if code != 2 {
		return fail("post_check_failed", "Deletion submitted but zonegroup absence was not verified")
	}
	after, afterDef, ok := list("list_after")
	if !ok || afterDef != def || !reflect.DeepEqual(after, remaining) {
		return fail("post_check_failed", "Deletion submitted but name index or current-context default changed unexpectedly")
	}
	if realm != "" {
		commitReq := Request{Action: req.Action + ".period", Parameters: map[string]any{"realm_id": realm, "expected_current_period": current}}
		commitSpec := command{args: []string{"period", "update", "--commit", "--realm-id", realm, "--format", "json"}, check: []string{"period", "get", "--realm-id", realm, "--format", "json"}}
		if _, err := s.executePeriodCommit(ctx, access, commitReq, commitSpec); err != nil {
			return fail("post_check_failed", "Zonegroup removed locally but Period publication failed or is uncertain; inspect partial state")
		}
		published, _, ok := run("published_absence", commitSpec.check, false)
		if !ok || !zonegroupAbsentFromPeriod(published, realm, id) {
			return fail("post_check_failed", "Published Period does not verify zonegroup absence")
		}
	}
	return cephdomain.ActionResult{Details: map[string]any{"zonegroup_id": id, "zonegroup_absence_verified": true, "period_published": realm != "", "zones_deleted": false, "pools_deleted": false}}, nil
}

func zonegroupAbsentFromPeriod(period map[string]any, realm, id string) bool {
	if period["realm_id"] != realm || !syncFlowToken(syncGroupString(period, "id")) {
		return false
	}
	pm, _ := period["period_map"].(map[string]any)
	groups, ok := pm["zonegroups"].([]any)
	if !ok || len(groups) == 0 {
		return false
	}
	seen := map[string]bool{}
	master := syncGroupString(period, "master_zonegroup")
	for _, raw := range groups {
		group, ok := raw.(map[string]any)
		gid := syncGroupString(group, "id")
		if !ok || !syncFlowToken(gid) || gid == id || seen[gid] {
			return false
		}
		seen[gid] = true
	}
	return master != "" && seen[master]
}
