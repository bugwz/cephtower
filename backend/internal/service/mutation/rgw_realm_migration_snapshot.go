package mutation

import "reflect"

// Migration retains the existing default zone rather than creating replacement
// pools. Keep unknown fields in the comparison so new native storage settings
// cannot silently disappear during a migration.
type realmMigrationSnapshot struct {
	group  map[string]any
	zone   map[string]any
	member map[string]any
}

func newRealmMigrationSnapshot(group, zone map[string]any, groupID, zoneID string) (realmMigrationSnapshot, bool) {
	if !syncFlowToken(groupID) || !syncFlowToken(zoneID) || group["id"] != groupID || zone["id"] != zoneID || group["name"] != "default" || zone["name"] != "default" || group["realm_id"] != "" || zone["realm_id"] != "" {
		return realmMigrationSnapshot{}, false
	}
	zones, ok := group["zones"].([]any)
	if !ok || len(zones) != 1 {
		return realmMigrationSnapshot{}, false
	}
	member, ok := zones[0].(map[string]any)
	if !ok || member["id"] != zoneID || member["name"] != "default" || group["master_zone"] != zoneID {
		return realmMigrationSnapshot{}, false
	}
	placements, ok := zone["placement_pools"].([]any)
	if !ok || len(placements) == 0 || rawText(zone, "domain_root") == "" {
		return realmMigrationSnapshot{}, false
	}
	return realmMigrationSnapshot{group: migrationUnchangedFields(group, "name", "realm_id", "is_master", "endpoints", "zones"), zone: migrationUnchangedFields(zone, "name", "realm_id", "system_key"), member: migrationUnchangedFields(member, "name", "endpoints", "tier_type")}, true
}

func migrationUnchangedFields(doc map[string]any, changed ...string) map[string]any {
	// Clone the whole native document, including nested placement/pool maps.
	clone := cloneMigrationDocument(doc)
	for _, field := range changed {
		delete(clone, field)
	}
	return clone
}

func cloneMigrationDocument(doc map[string]any) map[string]any {
	clone := make(map[string]any, len(doc))
	for key, value := range doc {
		clone[key] = cloneMigrationValue(value)
	}
	return clone
}

func cloneMigrationValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneMigrationDocument(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneMigrationValue(item)
		}
		return out
	default:
		return value
	}
}

func (snapshot realmMigrationSnapshot) storagePreserved(group, zone map[string]any) bool {
	zones, ok := group["zones"].([]any)
	if !ok || len(zones) != 1 {
		return false
	}
	member, ok := zones[0].(map[string]any)
	if !ok || !reflect.DeepEqual(snapshot.member, migrationUnchangedFields(member, "name", "endpoints", "tier_type")) {
		return false
	}
	return snapshot.group != nil && snapshot.zone != nil &&
		reflect.DeepEqual(snapshot.group, migrationUnchangedFields(group, "name", "realm_id", "is_master", "endpoints", "zones")) &&
		reflect.DeepEqual(snapshot.zone, migrationUnchangedFields(zone, "name", "realm_id", "system_key"))
}
