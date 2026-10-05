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

func buildZonegroupStorageClass(p map[string]any) (command, error) {
	for _, key := range []string{"name", "zonegroup_id", "placement_id", "storage_class"} {
		if !syncFlowToken(syncGroupString(p, key)) {
			return command{}, invalid("explicit group, target and new class required")
		}
	}
	realm, ok := p["realm_id"].(string)
	if _, valid := p["expected_default_placement"].(string); !valid {
		return command{}, invalid("expected default placement required")
	}
	if !ok || realm != "" && !syncFlowToken(realm) || strings.Contains(syncGroupString(p, "placement_id"), "/") || p["confirm_create"] != true {
		return command{}, invalid("explicit Realm, unqualified placement ID and confirmation required")
	}
	return command{binary: executor.BinaryRGWAdmin, args: []string{"zonegroup", "placement", "add", "--zonegroup-id", syncGroupString(p, "zonegroup_id"), "--placement-id", syncGroupString(p, "placement_id"), "--storage-class", syncGroupString(p, "storage_class"), "--format", "json"}}, nil
}

func (s *Service) executeZonegroupStorageClass(ctx context.Context, access executor.ClusterAccess, req Request, spec command) (cephdomain.ActionResult, error) {
	fail := func(code string) (cephdomain.ActionResult, error) {
		if req.Action == "rgw_zonegroup.storage_class_delete" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: "storage class removal or Period publication could not be verified; inspect local and published configuration before retrying; no Zone pool mapping or objects are removed", Retryable: false}
		}
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: "Zonegroup placement configuration could not be verified; inspect local configuration before retrying; no Period publication or Zone pool configuration performed", Retryable: false}
	}
	run := func(stage string, args []string, write bool) (map[string]any, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: req.Action + "." + stage, Binary: executor.BinaryRGWAdmin, Args: args, Mutating: write, Timeout: time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		return periodDocument(result.Stdout), err == nil && result.ExitCode == 0
	}
	p := req.Parameters
	args := []string{"zonegroup", "get", "--zonegroup-id", syncGroupString(p, "zonegroup_id"), "--format", "json"}
	before, ok := run("before", args, false)
	if !ok || before["id"] != p["zonegroup_id"] || before["name"] != p["name"] || before["realm_id"] != p["realm_id"] || before["default_placement"] != p["expected_default_placement"] {
		return fail("pre_check_failed")
	}
	createTarget := req.Action == "rgw_zonegroup.placement_create"
	var expected map[string]any
	if req.Action == "rgw_zonegroup.storage_class_delete" {
		expected, ok = zonegroupStorageClassDeleteExpected(before, p)
	} else if req.Action == "rgw_zonegroup.placement_tags" {
		expected, ok = zonegroupPlacementTagsExpected(before, p)
	} else if req.Action == "rgw_zonegroup.placement_default" {
		expected, ok = zonegroupPlacementDefaultExpected(before, p)
	} else if createTarget {
		expected, ok = zonegroupPlacementCreateExpected(before, p)
	} else {
		expected, ok = zonegroupStorageClassExpected(before, p)
	}
	if !ok {
		return fail("pre_check_failed")
	}
	realm := syncGroupString(p, "realm_id")
	current := ""
	if req.Action == "rgw_zonegroup.storage_class_delete" && realm != "" {
		r, valid := run("realm_before", []string{"realm", "get", "--realm-id", realm, "--format", "json"}, false)
		current = syncGroupString(r, "current_period")
		if !valid || r["id"] != realm || !syncFlowToken(current) {
			return fail("pre_check_failed")
		}
	}
	checked, ok := run("recheck", args, false)
	if !ok || !reflect.DeepEqual(checked, before) {
		return fail("pre_check_failed")
	}
	// Native placement commands emit a target map, not the whole group.
	// Use an independent get, including for default changes absent from that map.
	writeStage := "add"
	if req.Action == "rgw_zonegroup.storage_class_delete" {
		writeStage = "remove"
	}
	if _, ok := run(writeStage, spec.args, true); !ok {
		return fail("command_failed")
	}
	after, ok := run("after", args, false)
	if !ok || !reflect.DeepEqual(after, expected) {
		return fail("post_check_failed")
	}
	if req.Action == "rgw_zonegroup.storage_class_delete" {
		if realm != "" {
			commitReq := Request{Action: req.Action + ".period", Parameters: map[string]any{"realm_id": realm, "expected_current_period": current}}
			commitSpec := command{args: []string{"period", "update", "--commit", "--realm-id", realm, "--format", "json"}, check: []string{"period", "get", "--realm-id", realm, "--format", "json"}}
			if _, err := s.executePeriodCommit(ctx, access, commitReq, commitSpec); err != nil {
				return fail("post_check_failed")
			}
		}
		return cephdomain.ActionResult{Details: map[string]any{"zonegroup_id": p["zonegroup_id"], "placement_id": p["placement_id"], "storage_class": p["storage_class"], "class_removal_verified": true, "period_published": realm != "", "zone_mappings_removed": false, "objects_removed": false}}, nil
	}
	class := p["storage_class"]
	if createTarget {
		class = "STANDARD"
	}
	if req.Action == "rgw_zonegroup.placement_tags" {
		return cephdomain.ActionResult{Details: map[string]any{"zonegroup_id": p["zonegroup_id"], "placement_id": p["placement_id"], "tags_verified": true, "period_published": false, "default_placement_initialized": before["default_placement"] == ""}}, nil
	}
	if req.Action == "rgw_zonegroup.placement_default" {
		return cephdomain.ActionResult{Details: map[string]any{"zonegroup_id": p["zonegroup_id"], "default_placement": expected["default_placement"], "default_placement_verified": true, "period_published": false, "data_migrated": false}}, nil
	}
	return cephdomain.ActionResult{Details: map[string]any{"zonegroup_id": p["zonegroup_id"], "placement_id": p["placement_id"], "storage_class": class, "placement_created": createTarget, "declaration_verified": true, "period_published": false, "zone_pools_configured": false, "default_placement_initialized": before["default_placement"] == ""}}, nil
}

func zonegroupStorageClassExpected(group, p map[string]any) (map[string]any, bool) {
	raw, err := json.Marshal(group)
	if err != nil {
		return nil, false
	}
	expected := periodDocument(raw)
	clear(raw)
	def, ok := expected["default_placement"].(string)
	if !ok {
		return nil, false
	}
	targets, ok := expected["placement_targets"].([]any)
	if !ok {
		return nil, false
	}
	found := false
	seen := map[string]bool{}
	for _, raw := range targets {
		target, ok := raw.(map[string]any)
		name := syncGroupString(target, "name")
		if !ok || !syncFlowToken(name) || seen[name] {
			return nil, false
		}
		seen[name] = true
		if name != p["placement_id"] {
			continue
		}
		classes, ok := realmSetupStrings(target["storage_classes"])
		if !ok {
			return nil, false
		}
		for _, class := range classes {
			if class == p["storage_class"] {
				return nil, false
			}
		}
		if rawTiers, present := target["tier_targets"]; present {
			tiers, ok := rawTiers.([]any)
			if !ok {
				return nil, false
			}
			for _, rawTier := range tiers {
				tier, ok := rawTier.(map[string]any)
				if !ok || !syncFlowToken(syncGroupString(tier, "key")) || tier["key"] == p["storage_class"] {
					return nil, false
				}
			}
		}
		classes = append(classes, syncGroupString(p, "storage_class"))
		sort.Strings(classes)
		values := make([]any, len(classes))
		for i, class := range classes {
			values[i] = class
		}
		target["storage_classes"] = values
		found = true
	}
	if !found {
		return nil, false
	}
	if def == "" {
		expected["default_placement"] = p["placement_id"]
	}
	return expected, true
}
