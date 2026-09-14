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
