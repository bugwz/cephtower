package mutation

import (
	"reflect"
	"slices"
	"strings"
)

type migrationAdmin func(string, bool, ...string) (map[string]any, bool)
type migrationRename func(string, ...string) bool

func buildRealmMigration(p map[string]any) (command, error) {
	if p["confirm_migration"] != true || !syncFlowToken(rawText(p, "expected_zonegroup_id")) || !syncFlowToken(rawText(p, "expected_zone_id")) || p["zonegroup"] == "default" || p["zone"] == "default" {
		return command{}, invalid("migration requires explicit confirmation and default topology identities")
	}
	return buildRealmSetup(p)
}

// Keep rename's empty output separate from JSON-returning administrative calls.
// Native zone rename can return success without updating its group membership.
func migrateRealmTopology(admin migrationAdmin, rename migrationRename, snapshot realmMigrationSnapshot, p map[string]any, realmID, groupID, zoneID string) (map[string]any, map[string]any, bool) {
	groupName, zoneName := rawText(p, "zonegroup"), rawText(p, "zone")
	fail := func() (map[string]any, map[string]any, bool) { return nil, nil, false }
	if !rename("group_rename", "zonegroup", "rename", "--zonegroup-id", groupID, "--zonegroup-new-name", groupName) {
		return fail()
	}
	group, ok := admin("group_renamed", false, "zonegroup", "get", "--zonegroup-id", groupID)
	if !ok || group["id"] != groupID || group["name"] != groupName {
		return fail()
	}
	if !rename("zone_rename", "zone", "rename", "--zone-id", zoneID, "--zone-new-name", zoneName, "--zonegroup-id", groupID) {
		return fail()
	}
	zone, ok := admin("zone_renamed", false, "zone", "get", "--zone-id", zoneID)
	if !ok || zone["id"] != zoneID || zone["name"] != zoneName {
		return fail()
	}
	group, ok = admin("membership_renamed", false, "zonegroup", "get", "--zonegroup-id", groupID)
	if !ok || !migrationMemberMatches(group, zoneID, zoneName) || !snapshot.storagePreserved(group, zone) {
		return fail()
	}
	groupEP, ok := realmSetupStrings(p["zonegroup_endpoints"])
	if !ok {
		return fail()
	}
	zoneEP, ok := realmSetupStrings(p["zone_endpoints"])
	if !ok {
		return fail()
	}
	if _, ok := admin("group_migrate", true, "zonegroup", "modify", "--zonegroup-id", groupID, "--realm-id", realmID, "--master", "--default", "--endpoints", strings.Join(groupEP, ",")); !ok {
		return fail()
	}
	args := []string{"zone", "modify", "--zone-id", zoneID, "--zonegroup-id", groupID, "--realm-id", realmID, "--master", "--default", "--endpoints", strings.Join(zoneEP, ",")}
	if p["tier_type"] == "archive" {
		args = append(args, "--tier-type", "archive")
	}
	if _, ok := admin("zone_migrate", true, args...); !ok {
		return fail()
	}
	zone, ok = admin("zone_migrated", false, "zone", "get", "--zone-id", zoneID)
	if !ok || zone["id"] != zoneID || zone["name"] != zoneName || zone["realm_id"] != realmID {
		return fail()
	}
	group, ok = admin("group_migrated", false, "zonegroup", "get", "--zonegroup-id", groupID)
	if !ok || group["id"] != groupID || group["name"] != groupName || group["realm_id"] != realmID || group["is_master"] != true || !migrationMemberMatches(group, zoneID, zoneName) || !snapshot.storagePreservedAfterModify(group, zone) {
		return fail()
	}
	member := group["zones"].([]any)[0].(map[string]any)
	for _, pair := range []struct {
		actual   any
		expected []string
	}{{group["endpoints"], groupEP}, {member["endpoints"], zoneEP}} {
		actual, valid := realmSetupStrings(pair.actual)
		if !valid {
			return fail()
		}
		slices.Sort(actual)
		slices.Sort(pair.expected)
		if !reflect.DeepEqual(actual, pair.expected) {
			return fail()
		}
	}
	if rawText(member, "tier_type") != rawText(p, "tier_type") {
		return fail()
	}
	return group, zone, true
}

func migrationMemberMatches(group map[string]any, zoneID, name string) bool {
	zones, ok := group["zones"].([]any)
	if !ok || len(zones) != 1 || group["master_zone"] != zoneID {
		return false
	}
	member, ok := zones[0].(map[string]any)
	return ok && member["id"] == zoneID && member["name"] == name
}
