package mutation

import "testing"

func migrationSnapshotFixture() (map[string]any, map[string]any) {
	group := periodDocument([]byte(`{"id":"g","name":"default","realm_id":"","master_zone":"z","zones":[{"id":"z","name":"default"}],"placement_targets":[{"name":"default-placement","storage_classes":["STANDARD"]}],"default_placement":"default-placement","sync_policy":{"groups":[]}}`))
	zone := periodDocument([]byte(`{"id":"z","name":"default","realm_id":"","domain_root":"default.rgw.meta:root","placement_pools":[{"key":"default-placement","val":{"index_pool":"default.rgw.buckets.index","storage_classes":{"STANDARD":{"data_pool":"default.rgw.buckets.data"}}}}],"system_key":{"access_key":"old","secret_key":"old-secret"},"future_pool":"preserve-me"}`))
	member := group["zones"].([]any)[0].(map[string]any)
	member["log_data"], member["supported_features"] = false, []any{}
	return group, zone
}

func TestRealmMigrationNativeDerivedFields(t *testing.T) {
	g, z := migrationSnapshotFixture()
	member := g["zones"].([]any)[0].(map[string]any)
	member["log_data"], member["supported_features"] = true, []any{"old-feature"}
	snapshot, _ := newRealmMigrationSnapshot(g, z, "g", "z")
	member["log_data"], member["supported_features"] = false, []any{"old-feature", "new-native-feature"}
	if snapshot.storagePreserved(g, z) || !snapshot.storagePreservedAfterModify(g, z) {
		t.Fatal("native transitions not phase scoped")
	}
	for _, value := range []any{nil, []any{}, []any{"new-native-feature"}, []any{"old-feature", "old-feature"}, []any{true}} {
		member["supported_features"] = value
		if snapshot.storagePreservedAfterModify(g, z) {
			t.Fatal("invalid or removed feature accepted")
		}
	}
	member["supported_features"], member["log_data"] = []any{"old-feature"}, true
	if snapshot.storagePreservedAfterModify(g, z) {
		t.Fatal("incorrect singleton logging accepted")
	}
	member["log_data"], member["read_only"] = false, true
	if snapshot.storagePreservedAfterModify(g, z) {
		t.Fatal("unrelated member mutation accepted")
	}
}

func TestRealmMigrationSnapshotPreservesNativeStorage(t *testing.T) {
	g, z := migrationSnapshotFixture()
	snapshot, ok := newRealmMigrationSnapshot(g, z, "g", "z")
	if !ok {
		t.Fatal("default topology rejected")
	}
	g["name"], g["realm_id"], g["is_master"], g["endpoints"] = "new-group", "realm", true, []any{"https://gateway"}
	z["name"], z["realm_id"], z["system_key"] = "new-zone", "realm", map[string]any{"access_key": "new", "secret_key": "new-secret"}
	if !snapshot.storagePreserved(g, z) {
		t.Fatal("intended migration changes rejected")
	}
	z["placement_pools"].([]any)[0].(map[string]any)["val"].(map[string]any)["index_pool"] = "replacement"
	if snapshot.storagePreserved(g, z) {
		t.Fatal("nested pool change mutated or bypassed snapshot")
	}
	for _, field := range []string{"id", "domain_root", "placement_pools", "future_pool"} {
		g, z := migrationSnapshotFixture()
		snapshot, _ := newRealmMigrationSnapshot(g, z, "g", "z")
		delete(z, field)
		if snapshot.storagePreserved(g, z) {
			t.Fatal("removed field accepted", field)
		}
	}
	g, z = migrationSnapshotFixture()
	snapshot, _ = newRealmMigrationSnapshot(g, z, "g", "z")
	g["default_placement"] = "replacement"
	if snapshot.storagePreserved(g, z) {
		t.Fatal("group placement change accepted")
	}
}

func TestRealmMigrationSnapshotRejectsAmbiguousTopology(t *testing.T) {
	for _, mutate := range []func(map[string]any, map[string]any){
		func(g, z map[string]any) { g["realm_id"] = "existing" },
		func(g, z map[string]any) { z["realm_id"] = "existing" },
		func(g, z map[string]any) { g["zones"] = []any{} },
		func(g, z map[string]any) { g["master_zone"] = "other" },
		func(g, z map[string]any) { z["name"] = "other" },
		func(g, z map[string]any) { z["placement_pools"] = nil },
		func(g, z map[string]any) { g["zones"].([]any)[0].(map[string]any)["id"] = "other" },
	} {
		g, z := migrationSnapshotFixture()
		mutate(g, z)
		if _, ok := newRealmMigrationSnapshot(g, z, "g", "z"); ok {
			t.Fatal("ambiguous source accepted")
		}
	}
	if (realmMigrationSnapshot{}).storagePreserved(nil, nil) {
		t.Fatal("missing snapshot accepted")
	}
}
