package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

// Mirror RGWAM.zone_create without publishing an ordinary zone first. Every
// failure stops the chain; callers must report possible partial effects.
func (s *Service) importArchiveZone(ctx context.Context, access executor.ClusterAccess, request Request, token rgwRealmTokenDocument, originalSpec []byte) bool {
	name := rawText(request.Parameters, "name")
	run := func(stage string, binary executor.Binary, args []string, stdin []byte, mutating bool) ([]byte, bool) {
		sensitive := map[int]struct{}{}
		for i, arg := range args {
			if (arg == "--access-key" || arg == "--secret") && i+1 < len(args) {
				sensitive[i+1] = struct{}{}
			}
		}
		timeout := time.Minute
		if mutating {
			timeout = 10 * time.Minute
		}
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".archive." + stage, Binary: binary, Args: args, Stdin: stdin, Mutating: mutating, Timeout: timeout, MaxOutput: executor.DefaultMaxOutput, SensitiveArgs: sensitive})
		defer clear(result.Stderr)
		if err != nil || result.ExitCode != 0 {
			clear(result.Stdout)
			return nil, false
		}
		return result.Stdout, true
	}
	admin := func(stage string, mutating bool, args ...string) (map[string]any, bool) {
		raw, ok := run(stage, executor.BinaryRGWAdmin, append(args, "--format", "json"), nil, mutating)
		defer clear(raw)
		doc := periodDocument(raw)
		return doc, ok && doc != nil
	}
	pulled, ok := admin("pull", true, "realm", "pull", "--rgw-realm", token.RealmName, "--url", token.Endpoint, "--access-key", token.AccessKey, "--secret", token.Secret)
	if !ok || pulled["id"] != token.RealmID || pulled["name"] != token.RealmName || !syncFlowToken(rawText(pulled, "current_period")) {
		return false
	}
	period, ok := admin("period", false, "period", "get", "--realm-id", token.RealmID)
	if !ok || period["id"] != pulled["current_period"] || period["realm_id"] != token.RealmID {
		return false
	}
	group, master, ok := archiveMaster(period)
	if !ok || strings.ContainsAny(rawText(master, "name"), ",;= \t\n\r") {
		return false
	}
	// The new name must also be absent from the remote published topology.
	pm, _ := period["period_map"].(map[string]any)
	groups, _ := pm["zonegroups"].([]any)
	for _, item := range groups {
		g, ok := item.(map[string]any)
		if !ok {
			return false
		}
		zones, ok := g["zones"].([]any)
		if !ok {
			return false
		}
		for _, item := range zones {
			z, ok := item.(map[string]any)
			if !ok || z["name"] == name {
				return false
			}
		}
	}
	groupID, groupName := rawText(group, "id"), rawText(group, "name")
	zone, ok := admin("create", true, "zone", "create", "--realm-id", token.RealmID, "--zonegroup-id", groupID, "--rgw-zone", name, "--access-key", token.AccessKey, "--secret", token.Secret, "--tier-type", "archive", "--sync-from-all=false", "--sync-from", rawText(master, "name"))
	zoneID := rawText(zone, "id")
	if !ok || !syncFlowToken(zoneID) || zone["name"] != name || zone["realm_id"] != token.RealmID {
		return false
	}
	key, _ := zone["system_key"].(map[string]any)
	if key["access_key"] != token.AccessKey || key["secret_key"] != token.Secret {
		return false
	}
	actualGroup, ok := admin("group", false, "zonegroup", "get", "--zonegroup-id", groupID)
	if !ok || actualGroup["id"] != groupID || actualGroup["name"] != groupName || actualGroup["realm_id"] != token.RealmID || actualGroup["master_zone"] != master["id"] || !archiveZoneInGroup(actualGroup, zoneID, name, rawText(master, "name")) {
		return false
	}
	_, actualMaster, ok := archiveMaster(map[string]any{"master_zonegroup": groupID, "period_map": map[string]any{"zonegroups": []any{actualGroup}}})
	if !ok || actualMaster["name"] != master["name"] {
		return false
	}
	committed, ok := admin("commit", true, "period", "update", "--commit", "--realm-id", token.RealmID, "--zonegroup-id", groupID, "--zone-id", zoneID)
	current := rawText(committed, "id")
	if !ok || !syncFlowToken(current) || !archiveZonePublished(committed, token.RealmID, current, zoneID, name) {
		return false
	}
	committedGroup, committedMaster, ok := archiveMaster(committed)
	if !ok || committedGroup["id"] != groupID || committedGroup["name"] != groupName || committedMaster["id"] != master["id"] || committedMaster["name"] != master["name"] {
		return false
	}
	realm, ok := admin("realm_check", false, "realm", "get", "--realm-id", token.RealmID)
	if !ok || realm["id"] != token.RealmID || realm["name"] != token.RealmName || realm["current_period"] != current {
		return false
	}
	readback, ok := admin("period_check", false, "period", "get", "--realm-id", token.RealmID)
	if !ok || !reflect.DeepEqual(readback, committed) {
		return false
	}
	// Match the native module's secondary token and update_endpoints behavior.
	secondary, err := json.Marshal(map[string]any{"realm_name": token.RealmName, "realm_id": token.RealmID, "endpoint": nil, "access_key": token.AccessKey, "secret": token.Secret})
	if err != nil {
		return false
	}
	defer clear(secondary)
	var spec map[string]any
	if json.Unmarshal(originalSpec, &spec) != nil {
		return false
	}
	spec["rgw_realm"] = token.RealmName
	spec["rgw_zonegroup"] = groupName
	spec["rgw_realm_token"] = base64.StdEncoding.EncodeToString(secondary)
	spec["update_endpoints"] = true
	body, err := json.Marshal(spec)
	if err != nil {
		return false
	}
	defer clear(body)
	raw, ok := run("deploy", executor.BinaryCeph, []string{"orch", "apply", "-i", "-"}, body, true)
	clear(raw)
	return ok
}

// Resolve one master group and one master zone by native IDs, not list order.
func archiveMaster(period map[string]any) (map[string]any, map[string]any, bool) {
	fail := func() (map[string]any, map[string]any, bool) { return nil, nil, false }
	masterID := rawText(period, "master_zonegroup")
	if !syncFlowToken(masterID) {
		return fail()
	}
	pm, _ := period["period_map"].(map[string]any)
	groups, ok := pm["zonegroups"].([]any)
	if !ok {
		return fail()
	}
	var selected map[string]any
	seen := map[string]bool{}
	for _, item := range groups {
		group, ok := item.(map[string]any)
		id := rawText(group, "id")
		if !ok || !syncFlowToken(id) || seen[id] {
			return fail()
		}
		seen[id] = true
		if id == masterID {
			selected = group
		}
	}
	if selected == nil || !syncFlowToken(rawText(selected, "name")) || !syncFlowToken(rawText(selected, "master_zone")) {
		return fail()
	}
	zones, ok := selected["zones"].([]any)
	if !ok {
		return fail()
	}
	seen, names := map[string]bool{}, map[string]bool{}
	var master map[string]any
	for _, item := range zones {
		zone, ok := item.(map[string]any)
		id, name := rawText(zone, "id"), rawText(zone, "name")
		if !ok || !syncFlowToken(id) || !syncFlowToken(name) || seen[id] || names[name] {
			return fail()
		}
		seen[id], names[name] = true, true
		if id == selected["master_zone"] {
			master = zone
		}
	}
	return selected, master, master != nil
}

func archiveZoneInGroup(group map[string]any, id, name, masterName string) bool {
	zones, ok := group["zones"].([]any)
	if !ok || group["master_zone"] == id {
		return false
	}
	count := 0
	for _, item := range zones {
		zone, ok := item.(map[string]any)
		if !ok {
			return false
		}
		if zone["id"] != id && zone["name"] != name {
			continue
		}
		count++
		sources, ok := zone["sync_from"].([]any)
		if zone["id"] != id || zone["name"] != name || zone["tier_type"] != "archive" || zone["sync_from_all"] != false || !ok || len(sources) != 1 || sources[0] != masterName {
			return false
		}
	}
	return count == 1
}

func archiveZonePublished(period map[string]any, realm, current, id, name string) bool {
	epoch, ok := period["epoch"].(json.Number)
	number, err := epoch.Int64()
	if !ok || err != nil || number <= 0 {
		return false
	}
	_, ok = realmImportPublishedZone(period, realm, current, id, name)
	if !ok {
		return false
	}
	group, master, ok := archiveMaster(period)
	return ok && archiveZoneInGroup(group, id, name, rawText(master, "name"))
}
