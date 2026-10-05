package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func migrationParameters() map[string]any {
	p := setupParameters()
	p["confirm_migration"], p["expected_zonegroup_id"], p["expected_zone_id"] = true, "g", "z"
	return p
}

func migrationResponses(t *testing.T) map[string]string {
	t.Helper()
	r := setupResponses(t)
	encode := func(doc map[string]any) string {
		raw, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	g, z := migrationSnapshotFixture()
	r["migration_group"], r["migration_zone"] = encode(g), encode(z)
	r["zonegroup_absence"] = `{"zonegroups":["default"],"default_info":"g"}`
	r["zone_absence"] = `{"zones":["default"],"default_info":"z"}`
	r["migration_user_absence"] = `[]`
	g["name"] = "group"
	r["group_renamed"] = encode(g)
	z["name"] = "primary"
	g["zones"].([]any)[0].(map[string]any)["name"] = "primary"
	r["zone_renamed"], r["membership_renamed"] = encode(z), encode(g)
	g["realm_id"], g["is_master"], g["endpoints"] = "r", true, []any{"https://group.example"}
	g["zones"].([]any)[0].(map[string]any)["endpoints"] = []any{"https://zone.example"}
	z["realm_id"] = "r"
	r["group_migrate"], r["group_migrated"], r["migration_group_preserved"] = encode(g), encode(g), encode(g)
	r["zone_migrate"], r["zone_migrated"], r["migration_zone_preserved"] = encode(z), encode(z), encode(z)
	z["system_key"] = map[string]any{"access_key": "generated-access", "secret_key": "generated-secret"}
	r["zone_check"], r["migration_final_group"] = encode(z), encode(g)
	return r
}

func TestRealmMigrationFullChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	e := &realmSetupExecutor{responses: migrationResponses(t)}
	s.executor = e
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.migrate", Parameters: migrationParameters()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Details.(map[string]any)["zone_id"] != "z" {
		t.Fatal("zone identity replaced")
	}
	for _, spec := range e.specs {
		if strings.HasSuffix(spec.ID, "group_create") || strings.HasSuffix(spec.ID, "zone_create") {
			t.Fatal("created replacement topology")
		}
	}
}

func TestRealmMigrationArchiveChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	r := migrationResponses(t)
	for _, stage := range []string{"group_migrated", "migration_group_preserved", "migration_final_group"} {
		doc := periodDocument([]byte(r[stage]))
		doc["zones"].([]any)[0].(map[string]any)["tier_type"] = "archive"
		raw, _ := json.Marshal(doc)
		r[stage] = string(raw)
	}
	for _, stage := range []string{"initial_commit", "commit", "period_check"} {
		r[stage] = strings.ReplaceAll(r[stage], `"tier_type":""`, `"tier_type":"archive"`)
	}
	e := &realmSetupExecutor{responses: r}
	s.executor = e
	p := migrationParameters()
	p["tier_type"] = "archive"
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.migrate", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	for _, spec := range e.specs {
		if strings.HasSuffix(spec.ID, ".zone_migrate") && !strings.Contains(strings.Join(spec.Args, " "), "--tier-type archive") {
			t.Fatal("archive tier not applied")
		}
	}
}

func TestRealmMigrationStopsAtNativeFailure(t *testing.T) {
	for _, stage := range []string{"migration_group", "migration_zone", "migration_user_absence", "group_rename", "group_renamed", "zone_rename", "zone_renamed", "membership_renamed", "group_migrate", "zone_migrate", "zone_migrated", "group_migrated", "migration_group_preserved", "migration_zone_preserved", "migration_final_group"} {
		t.Run(stage, func(t *testing.T) {
			s, _, cluster := newCephUserService(t)
			e := &realmSetupExecutor{responses: migrationResponses(t), fail: stage}
			s.executor = e
			if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.migrate", Parameters: migrationParameters()}); err == nil {
				t.Fatal("failure accepted")
			}
			// Preflight snapshot reads may finish before the shared validation.
			if stage != "migration_group" && stage != "migration_group_preserved" && !strings.HasSuffix(e.specs[len(e.specs)-1].ID, "."+stage) {
				t.Fatal("continued after failure")
			}
		})
	}
}

func TestRealmMigrationRequiresExplicitSource(t *testing.T) {
	for _, bad := range []map[string]any{{"confirm_migration": false}, {"expected_zone_id": ""}, {"zone": "default"}, {"zonegroup": "default"}} {
		p := migrationParameters()
		for k, v := range bad {
			p[k] = v
		}
		if _, err := buildRealmMigration(p); err == nil {
			t.Fatal("invalid migration accepted")
		}
	}
	for stage, bad := range map[string]string{"realm_absence": `{"realms":["existing"]}`, "zone_absence": `{"zones":["default"],"default_info":"wrong"}`, "migration_user_absence": `["sys"]`} {
		s, _, cluster := newCephUserService(t)
		r := migrationResponses(t)
		r[stage] = bad
		e := &realmSetupExecutor{responses: r}
		s.executor = e
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.migrate", Parameters: migrationParameters()}); err == nil {
			t.Fatal("unsafe preflight accepted")
		}
		for _, spec := range e.specs {
			if spec.Mutating {
				t.Fatal("write before migration preflight passed")
			}
		}
	}
}

func TestRealmMigrationRejectsIncompleteTransitionEvidenceBeforeWrites(t *testing.T) {
	for _, field := range []string{"supported_features", "log_data"} {
		for _, value := range []any{nil, "unknown", []any{"duplicate", "duplicate"}} {
			s, _, cluster := newCephUserService(t)
			r := migrationResponses(t)
			group := periodDocument([]byte(r["migration_group"]))
			member := group["zones"].([]any)[0].(map[string]any)
			if value == nil {
				delete(member, field)
			} else {
				member[field] = value
			}
			raw, err := json.Marshal(group)
			if err != nil {
				t.Fatal(err)
			}
			r["migration_group"] = string(raw)
			e := &realmSetupExecutor{responses: r}
			s.executor = e
			if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.migrate", Parameters: migrationParameters()}); err == nil {
				t.Fatal("incomplete transition evidence accepted")
			}
			for _, spec := range e.specs {
				if spec.Mutating {
					t.Fatal("wrote before transition evidence was verified")
				}
			}
		}
	}
}
