package mutation

import (
	"reflect"
	"testing"
)

func TestRealmEditCommands(t *testing.T) {
	for _, tc := range []struct {
		name, next string
		def        any
		wantError  bool
		followups  int
	}{
		{"east", "west", false, false, 0}, {"east", "west", true, false, 1},
		{"east", "east", true, false, 0}, {"east", "east", false, true, 0},
		{"east", "west", "true", true, 0}, {"east", "", false, true, 0},
	} {
		c, err := build(Request{Action: "rgw_realm.update"}, map[string]any{"name": tc.name, "new_name": tc.next, "default": tc.def})
		if (err != nil) != tc.wantError {
			t.Fatalf("%+v err=%v", tc, err)
		}
		if err != nil {
			continue
		}
		want := []string{"realm", "rename", "--rgw-realm", tc.name, "--realm-new-name", tc.next, "--format", "json"}
		if tc.name == tc.next {
			want = []string{"realm", "default", "--rgw-realm", tc.next, "--format", "json"}
		}
		if !reflect.DeepEqual(c.args, want) || len(c.followups) != tc.followups {
			t.Fatalf("unexpected command %+v", c)
		}
		if !reflect.DeepEqual(c.check, []string{"realm", "get", "--rgw-realm", tc.next, "--format", "json"}) {
			t.Fatal("readback uses old name")
		}
		if tc.followups > 0 && !reflect.DeepEqual(c.followups[0].args, []string{"realm", "default", "--rgw-realm", tc.next, "--format", "json"}) {
			t.Fatal("default uses old name")
		}
	}
}

func TestRealmCreateDefault(t *testing.T) {
	for _, value := range []any{nil, false, true, "true"} {
		parameters := map[string]any{"name": "east"}
		if value != nil {
			parameters["default"] = value
		}
		c, err := build(Request{Action: "rgw_realm.create"}, parameters)
		if value == "true" {
			if err == nil {
				t.Fatal("string boolean accepted")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"realm", "create", "--rgw-realm", "east"}
		if value == true {
			want = append(want, "--default")
		}
		want = append(want, "--format", "json")
		if !reflect.DeepEqual(c.args, want) {
			t.Fatalf("args=%v want=%v", c.args, want)
		}
	}
}

func TestZonegroupCreateOptions(t *testing.T) {
	p := map[string]any{"name": "east", "realm": "global", "master": true, "default": true, "endpoints": "https://a.example,https://b.example"}
	c, err := build(Request{Action: "rgw_zonegroup.create"}, p)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"zonegroup", "create", "--rgw-zonegroup", "east", "--default", "--master", "--rgw-realm", "global", "--endpoints", "https://a.example,https://b.example", "--format", "json"}
	if !reflect.DeepEqual(c.args, want) {
		t.Fatalf("args=%v", c.args)
	}
	for _, field := range []string{"master", "default", "realm", "endpoints"} {
		bad := map[string]any{"name": "east", field: 42}
		if _, err := build(Request{Action: "rgw_zonegroup.create"}, bad); err == nil {
			t.Fatalf("accepted invalid %s", field)
		}
	}
}

func TestZonegroupEditCommitsScopedPeriod(t *testing.T) {
	c, err := build(Request{Action: "rgw_zonegroup.update"}, map[string]any{"name": "east", "new_name": "west", "realm_id": "realm-id", "endpoints": "https://rgw.example"})
	if err != nil {
		t.Fatal(err)
	}
	if c.args[1] != "rename" || len(c.followups) != 2 || c.followups[0].args[1] != "modify" {
		t.Fatalf("wrong sequence %+v", c)
	}
	if !reflect.DeepEqual(c.followups[1].args, []string{"period", "update", "--commit", "--realm-id", "realm-id", "--format", "json"}) {
		t.Fatal("period scope missing")
	}
	if !reflect.DeepEqual(c.followups[1].check, []string{"zonegroup", "get", "--rgw-zonegroup", "west", "--format", "json"}) {
		t.Fatal("wrong readback")
	}
	for _, p := range []map[string]any{{"name": "east", "new_name": "east", "realm_id": "id"}, {"name": "east", "new_name": "east", "realm_id": "id", "master": "true"}} {
		if _, err := build(Request{Action: "rgw_zonegroup.update"}, p); err == nil {
			t.Fatalf("accepted invalid parameters %v", p)
		}
	}
}

func TestStandaloneZonegroupEdit(t *testing.T) {
	c, err := build(Request{Action: "rgw_zonegroup.update"}, map[string]any{"name": "east", "new_name": "west", "realm_id": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.followups) != 1 || c.followups[0].args[1] != "modify" || len(c.followups[0].check) == 0 {
		t.Fatalf("unexpected standalone commands %+v", c)
	}
	for _, cmd := range append([]command{c}, c.followups...) {
		for _, arg := range cmd.args {
			if arg == "--realm-id" || arg == "period" {
				t.Fatal("standalone edit must not change realm or commit a period")
			}
		}
	}
}

func TestZonegroupMembership(t *testing.T) {
	p := map[string]any{"name": "east", "new_name": "west", "realm_id": "realm", "add_zones": []any{"z1"}, "remove_zones": []any{"z2"}}
	c, err := build(Request{Action: "rgw_zonegroup.update"}, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.followups) != 4 {
		t.Fatalf("followups=%v", c.followups)
	}
	for i, action := range []string{"add", "remove"} {
		want := []string{"zonegroup", action, "--rgw-zonegroup", "west", "--rgw-zone", []string{"z1", "z2"}[i], "--format", "json"}
		if !reflect.DeepEqual(c.followups[i+1].args, want) {
			t.Fatalf("membership command=%v", c.followups[i+1].args)
		}
	}
	if c.followups[3].args[0] != "period" {
		t.Fatal("period must follow membership changes")
	}
	for _, bad := range []any{[]any{42}, "z1", []any{"z1", "z1"}, []any{""}, []any{"z2"}} {
		p["add_zones"] = bad
		if _, err := build(Request{Action: "rgw_zonegroup.update"}, p); err == nil {
			t.Fatalf("accepted %v", bad)
		}
	}
}

func TestZoneCreateTopology(t *testing.T) {
	c, err := build(Request{Action: "rgw_zone.create"}, map[string]any{"name": "zone-a", "zonegroup": "group-a", "endpoints": "https://rgw.example", "master": true, "default": true})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"zone", "create", "--rgw-zone", "zone-a", "--default", "--master", "--rgw-zonegroup", "group-a", "--endpoints", "https://rgw.example", "--format", "json"}
	if !reflect.DeepEqual(c.args, want) {
		t.Fatalf("args=%v", c.args)
	}
	for _, field := range []string{"zonegroup", "endpoints", "master", "default"} {
		if _, err := build(Request{Action: "rgw_zone.create"}, map[string]any{"name": "zone-a", field: 42}); err == nil {
			t.Fatalf("accepted invalid %s", field)
		}
	}
}

func TestZoneSyncOptions(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		c, err := build(Request{Action: "rgw_zone.create"}, map[string]any{"name": "east", "sync_from_all": enabled, "sync_from": "z1,z2"})
		if err != nil {
			t.Fatal(err)
		}
		flag := "--sync-from-all=false"
		if enabled {
			flag = "--sync-from-all=true"
		}
		want := []string{"zone", "create", "--rgw-zone", "east", flag, "--sync-from", "z1,z2", "--format", "json"}
		if !reflect.DeepEqual(c.args, want) {
			t.Fatalf("args=%v", c.args)
		}
	}
	for _, p := range []map[string]any{{"name": "east", "sync_from_all": "false"}, {"name": "east", "sync_from": 42}, {"name": "east", "sync_from": ""}} {
		if _, err := build(Request{Action: "rgw_zone.create"}, p); err == nil {
			t.Fatalf("accepted %v", p)
		}
	}
}

func TestArchiveZoneCreation(t *testing.T) {
	c, err := build(Request{Action: "rgw_zone.create"}, map[string]any{"name": "archive-a", "tier_type": "archive"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"zone", "create", "--rgw-zone", "archive-a", "--tier-type", "archive", "--format", "json"}
	if !reflect.DeepEqual(c.args, want) {
		t.Fatalf("args=%v", c.args)
	}
	for _, value := range []any{true, "", "unsupported"} {
		if _, err := build(Request{Action: "rgw_zone.create"}, map[string]any{"name": "a", "tier_type": value}); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
}

func TestZoneCredentialsAreSensitive(t *testing.T) {
	c, err := build(Request{Action: "rgw_zone.create"}, map[string]any{"name": "east", "access_key": "test-access", "secret_key": "test-secret"})
	if err != nil {
		t.Fatal(err)
	}
	for i, arg := range c.args {
		if arg == "test-access" || arg == "test-secret" {
			if _, ok := c.sensitive[i]; !ok {
				t.Fatal("credential argument not marked sensitive")
			}
		}
	}
	if len(c.sensitive) != 2 {
		t.Fatal("missing sensitive arguments")
	}
	for _, p := range []map[string]any{{"name": "east", "access_key": "a"}, {"name": "east", "secret_key": "b"}, {"name": "east", "access_key": "a", "secret_key": ""}} {
		if _, err := build(Request{Action: "rgw_zone.create"}, p); err == nil {
			t.Fatal("accepted incomplete credentials")
		}
	}
}
