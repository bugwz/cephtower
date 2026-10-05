package mutation

import (
	"reflect"
	"testing"
)

func TestRealmMigrationTopologySequence(t *testing.T) {
	stages := []string{"group_rename", "group_renamed", "zone_rename", "zone_renamed", "membership_renamed", "group_migrate", "zone_migrate", "zone_migrated", "group_migrated"}
	for _, failure := range append([]string{"", "membership_missing", "pool_changed"}, stages...) {
		t.Run(failure, func(t *testing.T) {
			group, zone := migrationSnapshotFixture()
			snapshot, _ := newRealmMigrationSnapshot(group, zone, "g", "z")
			var calls []string
			admin := func(stage string, write bool, args ...string) (map[string]any, bool) {
				calls = append(calls, stage)
				if stage == failure {
					return nil, false
				}
				if stage == "group_migrate" {
					group["realm_id"], group["is_master"] = "r", true
					group["endpoints"] = []any{"https://group"}
				}
				if stage == "zone_migrate" {
					zone["realm_id"] = "r"
					group["zones"].([]any)[0].(map[string]any)["endpoints"] = []any{"https://zone"}
				}
				if stage == "zone_renamed" && failure == "pool_changed" {
					zone["domain_root"] = "new-pool"
				}
				if args[0] == "zone" {
					return cloneMigrationDocument(zone), true
				}
				return cloneMigrationDocument(group), true
			}
			rename := func(stage string, args ...string) bool {
				calls = append(calls, stage)
				if stage == failure {
					return false
				}
				if stage == "group_rename" {
					if !reflect.DeepEqual(args, []string{"zonegroup", "rename", "--zonegroup-id", "g", "--zonegroup-new-name", "new-group"}) {
						t.Fatal("unscoped group rename")
					}
					group["name"] = "new-group"
				} else {
					if !reflect.DeepEqual(args, []string{"zone", "rename", "--zone-id", "z", "--zone-new-name", "new-zone", "--zonegroup-id", "g"}) {
						t.Fatal("unscoped zone rename")
					}
					zone["name"] = "new-zone"
					if failure != "membership_missing" {
						group["zones"].([]any)[0].(map[string]any)["name"] = "new-zone"
					}
				}
				return true
			}
			_, _, ok := migrateRealmTopology(admin, rename, snapshot, map[string]any{"zonegroup": "new-group", "zone": "new-zone", "zonegroup_endpoints": []string{"https://group"}, "zone_endpoints": []string{"https://zone"}}, "r", "g", "z")
			if ok != (failure == "") {
				t.Fatal("unexpected migration result")
			}
			if failure == "" && !reflect.DeepEqual(calls, stages) {
				t.Fatal("incorrect sequence", calls)
			}
			if failure == "membership_missing" || failure == "pool_changed" {
				if calls[len(calls)-1] != "membership_renamed" {
					t.Fatal("continued after invalid rename")
				}
			} else if failure != "" && calls[len(calls)-1] != failure {
				t.Fatal("continued after failure")
			}
		})
	}
}
