package clusterinspect

import (
	"context"
	"reflect"
	"testing"
)

func TestSubvolumeSnapshotVisibility(t *testing.T) {
	for _, output := range []string{"0", "1\n", "true", `"1"`, "", "1 0", "2"} {
		t.Run(output, func(t *testing.T) {
			s, runner, id := testInspection(t)
			runner.output = output
			result, err := s.SubvolumeSnapshotVisibility(context.Background(), id, "cephfs", "home", "users")
			valid := output == "0" || output == "1\n"
			if valid != (err == nil) {
				t.Fatalf("result=%+v error=%v", result, err)
			}
			if valid && (result.Visible != (output == "1\n") || result.Group != "users" || result.ObservedAt.IsZero()) {
				t.Fatalf("result=%+v", result)
			}
			want := []string{"fs", "subvolume", "snapshot_visibility", "get", "cephfs", "home", "--group_name", "users"}
			if len(runner.specs) != 1 || !reflect.DeepEqual(runner.specs[0].Args, want) || runner.specs[0].Mutating {
				t.Fatalf("specs=%+v", runner.specs)
			}
		})
	}
	s, runner, id := testInspection(t)
	runner.output = "1"
	if _, err := s.SubvolumeSnapshotVisibility(context.Background(), id, "cephfs", "home", "_nogroup"); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs[0].Args) != 6 {
		t.Fatal(runner.specs)
	}
	for _, scope := range [][3]string{{"--bad", "home", ""}, {"cephfs", "../home", ""}, {"cephfs", "home", "group\n"}} {
		before := len(runner.specs)
		if _, err := s.SubvolumeSnapshotVisibility(context.Background(), id, scope[0], scope[1], scope[2]); err == nil || len(runner.specs) != before {
			t.Fatalf("invalid scope=%v", scope)
		}
	}
	runner.fail = true
	if _, err := s.SubvolumeSnapshotVisibility(context.Background(), id, "cephfs", "home", ""); err == nil {
		t.Fatal("native failure accepted")
	}
}
