package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	cephintegration "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func buildZonePlacement(p map[string]any) (command, error) {
	for _, key := range []string{"zone_id", "name", "zonegroup_id", "placement_id", "storage_class"} {
		if !syncFlowToken(syncGroupString(p, key)) {
			return command{}, invalid("explicit Zone, group, placement and class identities required")
		}
	}
	realm, ok := p["realm_id"].(string)
	if !ok || realm != "" && !syncFlowToken(realm) || p["confirm_placement"] != true {
		return command{}, invalid("explicit Realm and placement confirmation required")
	}
	for _, key := range []string{"index_pool", "data_pool", "data_extra_pool"} {
		value, ok := p[key].(string)
		if !ok || value == "" && key != "data_extra_pool" || value != "" && !cephintegration.ValidRGWPoolReference(value) {
			return command{}, invalid("canonical pool references required")
		}
	}
	compression := syncGroupString(p, "compression")
	switch compression {
	case "none", "lz4", "zlib", "snappy", "zstd", "brotli":
	default:
		return command{}, invalid("unsupported compression algorithm")
	}
	args := []string{"zone", "placement", "modify", "--zone-id", syncGroupString(p, "zone_id"), "--zonegroup-id", syncGroupString(p, "zonegroup_id"), "--placement-id", syncGroupString(p, "placement_id"), "--storage-class", syncGroupString(p, "storage_class"), "--index-pool", syncGroupString(p, "index_pool"), "--data-pool", syncGroupString(p, "data_pool"), "--data-extra-pool", syncGroupString(p, "data_extra_pool"), "--compression", compression, "--format", "json"}
	return command{binary: executor.BinaryRGWAdmin, args: args}, nil
}

func (s *Service) executeZonePlacement(ctx context.Context, access executor.ClusterAccess, req Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: "Zone placement update could not be verified; inspect configuration before retrying; no data migration or gateway restart is performed", Retryable: false}
	}
	run := func(stage string, args []string, write bool) (map[string]any, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: req.Action + "." + stage, Binary: executor.BinaryRGWAdmin, Args: args, Mutating: write, Timeout: time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		return periodDocument(result.Stdout), err == nil && result.ExitCode == 0
	}
	p := req.Parameters
	realm := syncGroupString(p, "realm_id")
	zoneArgs := []string{"zone", "get", "--zone-id", syncGroupString(p, "zone_id"), "--format", "json"}
	groupArgs := []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id"), "--format", "json"}
	zone, ok := run("zone_before", zoneArgs, false)
	if !ok || zone["id"] != p["zone_id"] || zone["name"] != p["name"] || zone["realm_id"] != realm {
		return fail("pre_check_failed")
	}
	group, ok := run("group_before", groupArgs, false)
	if !ok || !zonePlacementGroupMatches(group, p) {
		return fail("pre_check_failed")
	}
	expected, ok := zonePlacementExpected(zone, p)
	if !ok {
		return fail("pre_check_failed")
	}
	current := ""
	if realm != "" {
		r, ok := run("realm_before", []string{"realm", "get", "--realm-id", realm, "--format", "json"}, false)
		current = syncGroupString(r, "current_period")
		if !ok || r["id"] != realm || !syncFlowToken(current) {
			return fail("pre_check_failed")
		}
	}
	checked, ok := run("group_recheck", groupArgs, false)
	if !ok || !reflect.DeepEqual(checked, group) {
		return fail("pre_check_failed")
	}
	checked, ok = run("zone_recheck", zoneArgs, false)
	if !ok || !reflect.DeepEqual(checked, zone) {
		return fail("pre_check_failed")
	}
	written, ok := run("modify", spec.args, true)
	if !ok {
		return fail("command_failed")
	}
	if !reflect.DeepEqual(written, expected) {
		return fail("post_check_failed")
	}
	checked, ok = run("zone_after", zoneArgs, false)
	if !ok || !reflect.DeepEqual(checked, expected) {
		return fail("post_check_failed")
	}
	checked, ok = run("group_after", groupArgs, false)
	if !ok || !reflect.DeepEqual(checked, group) {
		return fail("post_check_failed")
	}
	if realm != "" {
		commitReq := Request{Action: req.Action + ".period", Parameters: map[string]any{"realm_id": realm, "expected_current_period": current}}
		commitSpec := command{args: []string{"period", "update", "--commit", "--realm-id", realm, "--format", "json"}, check: []string{"period", "get", "--realm-id", realm, "--format", "json"}}
		if _, err := s.executePeriodCommit(ctx, access, commitReq, commitSpec); err != nil {
			return fail("post_check_failed")
		}
	}
	return cephdomain.ActionResult{Details: map[string]any{"zone_id": p["zone_id"], "placement_id": p["placement_id"], "storage_class": p["storage_class"], "placement_verified": true, "period_published": realm != "", "data_migrated": false}}, nil
}

func zonePlacementGroupMatches(group, p map[string]any) bool {
	if group["id"] != p["zonegroup_id"] || group["realm_id"] != p["realm_id"] {
		return false
	}
	members, ok := group["zones"].([]any)
	if !ok {
		return false
	}
	found := false
	seen := map[string]bool{}
	for _, raw := range members {
		m, ok := raw.(map[string]any)
		id := syncGroupString(m, "id")
		if !ok || !syncFlowToken(id) || seen[id] {
			return false
		}
		seen[id] = true
		found = found || id == p["zone_id"]
	}
	if !found {
		return false
	}
	targets, ok := group["placement_targets"].([]any)
	if !ok {
		return false
	}
	found = false
	seen = map[string]bool{}
	for _, raw := range targets {
		target, ok := raw.(map[string]any)
		name := syncGroupString(target, "name")
		if !ok || !syncFlowToken(name) || seen[name] {
			return false
		}
		seen[name] = true
		if name != p["placement_id"] {
			continue
		}
		classes, valid := realmSetupStrings(target["storage_classes"])
		if !valid {
			return false
		}
		for _, class := range classes {
			found = found || class == p["storage_class"]
		}
		tiers := []any{}
		if rawTiers, present := target["tier_targets"]; present {
			var valid bool
			tiers, valid = rawTiers.([]any)
			if !valid {
				return false
			}
		}
		for _, rawTier := range tiers {
			tier, ok := rawTier.(map[string]any)
			if !ok || !syncFlowToken(syncGroupString(tier, "key")) || tier["key"] == p["storage_class"] {
				return false
			}
		}
	}
	return found
}

func zonePlacementExpected(zone, p map[string]any) (map[string]any, bool) {
	raw, err := json.Marshal(zone)
	if err != nil {
		return nil, false
	}
	expected := periodDocument(raw)
	clear(raw)
	placements, ok := expected["placement_pools"].([]any)
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	found := false
	for _, raw := range placements {
		entry, ok := raw.(map[string]any)
		key := syncGroupString(entry, "key")
		if !ok || !syncFlowToken(key) || seen[key] {
			return nil, false
		}
		seen[key] = true
		if key != p["placement_id"] {
			continue
		}
		info, ok := entry["val"].(map[string]any)
		if !ok {
			return nil, false
		}
		classes, ok := info["storage_classes"].(map[string]any)
		if !ok {
			return nil, false
		}
		class, ok := classes[syncGroupString(p, "storage_class")].(map[string]any)
		if !ok {
			return nil, false
		}
		info["index_pool"], info["data_extra_pool"] = p["index_pool"], p["data_extra_pool"]
		class["data_pool"], class["compression_type"] = p["data_pool"], p["compression"]
		found = true
	}
	return expected, found
}
