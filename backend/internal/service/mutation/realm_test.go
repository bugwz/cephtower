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
