package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func groupUpdateRequest(params map[string]any) Request {
	return Request{Action: "subvolume_group.update", ResourceKey: "filesystem/cephfs/subvolume-group/users", Parameters: params}
}

func TestSubvolumeGroupUpdateCommands(t *testing.T) {
	attrs := map[string]any{"pool": "cephfs.hot", "uid": "1000", "gid": "1001", "mode": "0700"}
	request := groupUpdateRequest(attrs)
	cmd, err := build(request, attrs)
	wantAttrs := []string{"fs", "subvolumegroup", "create", "cephfs", "users", "--pool_layout", "cephfs.hot", "--uid", "1000", "--gid", "1001", "--mode", "0700"}
	if err != nil || !reflect.DeepEqual(cmd.args, wantAttrs) || len(cmd.followups) != 0 {
		t.Fatalf("command=%+v err=%v", cmd, err)
	}
	attrs["size"] = "1024"
	attrs["no_shrink"] = true
	cmd, err = build(request, attrs)
	wantResize := []string{"fs", "subvolumegroup", "resize", "cephfs", "users", "1024", "--no_shrink"}
	if err != nil || !reflect.DeepEqual(cmd.args, wantResize) || len(cmd.followups) != 1 || !reflect.DeepEqual(cmd.followups[0].args, wantAttrs) {
		t.Fatalf("command=%+v err=%v", cmd, err)
	}
	params := map[string]any{"unlimited": true}
	cmd, err = build(groupUpdateRequest(params), params)
	if err != nil || !reflect.DeepEqual(cmd.args, []string{"fs", "subvolumegroup", "resize", "cephfs", "users", "inf"}) {
		t.Fatalf("command=%+v err=%v", cmd, err)
	}
	for _, params := range []map[string]any{nil, {"size": "0"}, {"size": "-1"}, {"size": "1.5"}, {"size": "1024", "unlimited": true}, {"no_shrink": true}, {"pool": "hot"}, {"pool": "hot", "uid": "4294967296", "gid": "1", "mode": "0755"}, {"pool": "hot", "uid": "0", "gid": "0", "mode": "0899"}} {
		if _, err := build(groupUpdateRequest(params), params); err == nil {
			t.Fatalf("invalid params accepted=%v", params)
		}
	}
}

func TestSubvolumeGroupUpdateExecution(t *testing.T) {
	const before = `{"data_pool":"cephfs.cold","uid":0,"gid":0,"mode":16877,"bytes_quota":"infinite"}`
	const after = `{"data_pool":"cephfs.hot","uid":1000,"gid":1001,"mode":16832,"bytes_quota":1024}`
	for _, tt := range []struct {
		name, pre, post, failID, code string
		count                         int
	}{
		{"success", before, after, "", "", 4},
		{"group missing", before, after, "subvolume_group.update.pre_check", "ceph_command_failed", 1},
		{"invalid precheck", "{}", after, "", "invalid_request", 1},
		{"resize rejected", before, after, "subvolume_group.update", "ceph_command_failed", 2},
		{"attribute rejected", before, after, "subvolume_group.update.step2", "ceph_command_failed", 3},
		{"readback rejected", before, after, "subvolume_group.update.post_check", "post_check_failed", 4},
		{"unchanged", before, before, "", "post_check_failed", 4},
		{"malformed", before, "{}", "", "post_check_failed", 4},
		{"trailing response", before, after + ` {}`, "", "post_check_failed", 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"subvolume_group.update.pre_check": tt.pre, "subvolume_group.update.post_check": tt.post}, failID: tt.failID}
			s.executor = runner
			request := groupUpdateRequest(map[string]any{"pool": "cephfs.hot", "uid": "1000", "gid": "1001", "mode": "0700", "size": "1024", "no_shrink": true})
			request.ClusterID = id
			_, err := s.Execute(context.Background(), request)
			if tt.code == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != tt.code {
					t.Fatalf("err=%v want=%s", err, tt.code)
				}
			}
			if len(runner.specs) != tt.count {
				t.Fatalf("specs=%+v", runner.specs)
			}
		})
	}
}

func TestSubvolumeGroupReadbackFields(t *testing.T) {
	params := map[string]any{"pool": "hot", "uid": "1", "gid": "2", "mode": "2750", "unlimited": true}
	const valid = `{"data_pool":"hot","uid":1,"gid":2,"mode":17896,"bytes_quota":"infinite"}`
	if !subvolumeGroupUpdateMatches(params, []byte(valid)) {
		t.Fatal("valid permissions or unlimited quota rejected")
	}
	for _, output := range []string{`{"data_pool":"cold","uid":1,"gid":2,"mode":17896,"bytes_quota":"infinite"}`, `{"data_pool":"hot","uid":9,"gid":2,"mode":17896,"bytes_quota":"infinite"}`, `{"data_pool":"hot","uid":1,"gid":9,"mode":17896,"bytes_quota":"infinite"}`, `{"data_pool":"hot","uid":1,"gid":2,"mode":16877,"bytes_quota":"infinite"}`, `{"data_pool":"hot","uid":1,"gid":2,"mode":17896,"bytes_quota":1024}`} {
		if subvolumeGroupUpdateMatches(params, []byte(output)) {
			t.Fatalf("mismatch accepted=%s", output)
		}
	}
}

func TestSubvolumeGroupSingleDimensionUpdates(t *testing.T) {
	const info = `{"data_pool":"hot","uid":1,"gid":2,"mode":16877,"bytes_quota":"infinite"}`
	for _, tt := range []struct {
		params map[string]any
		count  int
	}{
		{map[string]any{"unlimited": true}, 2},
		{map[string]any{"pool": "hot", "uid": "1", "gid": "2", "mode": "0755"}, 3},
	} {
		s, _, id := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"subvolume_group.update.pre_check": info, "subvolume_group.update.post_check": info}}
		s.executor = runner
		request := groupUpdateRequest(tt.params)
		request.ClusterID = id
		if _, err := s.Execute(context.Background(), request); err != nil {
			t.Fatal(err)
		}
		if len(runner.specs) != tt.count {
			t.Fatalf("specs=%+v", runner.specs)
		}
	}
}
