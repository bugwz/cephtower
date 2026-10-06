package mutation

import (
	"reflect"
	"testing"
)

func TestOSDRemovalPreserveID(t *testing.T) {
	for _, preserve := range []bool{false, true} {
		for _, zap := range []bool{false, true} {
			spec, err := build(Request{Action: "osd.delete", ResourceKey: "osd/0"}, map[string]any{"preserve_id": preserve, "zap": zap})
			if err != nil {
				t.Fatal(err)
			}
			want := []string{"orch", "osd", "rm", "0"}
			if preserve {
				want = append(want, "--replace")
			}
			if zap {
				want = append(want, "--zap")
			}
			if !reflect.DeepEqual(spec.args, want) {
				t.Fatalf("args=%v want=%v", spec.args, want)
			}
			if !reflect.DeepEqual(spec.check, []string{"orch", "osd", "rm", "status", "--format", "json"}) {
				t.Fatal(spec.check)
			}
		}
	}
	for _, id := range []string{"all", "*", "--force", "01", "-1", "2147483648"} {
		if _, err := build(Request{Action: "osd.delete", ResourceKey: "osd/" + id}, map[string]any{"preserve_id": true}); err == nil {
			t.Fatal("invalid target accepted", id)
		}
	}
}
