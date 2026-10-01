package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestSubvolumeVisibilityCommandAndReadback(t *testing.T) {
	for _, visible := range []bool{true, false} {
		value, output, opposite := "true", "1\n", "0"
		if !visible {
			value, output, opposite = "false", "0", "1"
		}
		params := map[string]any{"visible": visible, "group": "users"}
		request := Request{Action: "subvolume.snapshot_visibility", ResourceKey: "filesystem/cephfs/subvolume/home", Parameters: params}
		cmd, err := build(request, params)
		want := []string{"fs", "subvolume", "snapshot_visibility", "set", "cephfs", "home", value, "--group_name", "users"}
		if err != nil || !reflect.DeepEqual(cmd.args, want) {
			t.Fatalf("command=%+v err=%v", cmd, err)
		}
		for _, readback := range []string{output, opposite, "2", "", "true", "1 0"} {
			s, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"subvolume.snapshot_visibility.post_check": readback}}
			s.executor = runner
			request.ClusterID = id
			_, err := s.Execute(context.Background(), request)
			if readback == output {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" {
					t.Fatalf("err=%v", err)
				}
			}
			if len(runner.specs) != 2 || runner.specs[1].Mutating {
				t.Fatal(runner.specs)
			}
		}
	}
	for _, params := range []map[string]any{nil, {"visible": "true"}, {"visible": true, "group": "../bad"}} {
		if _, err := build(Request{Action: "subvolume.snapshot_visibility", ResourceKey: "filesystem/cephfs/subvolume/home"}, params); err == nil {
			t.Fatalf("invalid params accepted=%v", params)
		}
	}
	cmd, err := build(Request{Action: "subvolume.snapshot_visibility", ResourceKey: "filesystem/cephfs/subvolume/home"}, map[string]any{"visible": false, "group": "_nogroup"})
	if err != nil || len(cmd.args) != 7 || len(cmd.check) != 6 {
		t.Fatalf("default group command=%+v err=%v", cmd, err)
	}
	cmd, err = build(Request{Action: "subvolume.snapshot_visibility", ResourceKey: "filesystem/subvolume/subvolume/group/group/users"}, map[string]any{"visible": true, "group": "users"})
	if err != nil || cmd.args[4] != "subvolume" || cmd.args[5] != "group" {
		t.Fatalf("reserved word name: command=%+v err=%v", cmd, err)
	}
}
