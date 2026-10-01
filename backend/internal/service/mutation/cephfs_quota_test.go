package mutation

import "testing"

func TestCephFSCreateQuotaExact(t *testing.T) {
	for _, action := range []string{"subvolume.create", "subvolume_group.create"} {
		for _, size := range []string{"0", "9007199254740993", "9223372036854775807"} {
			params := map[string]any{"name": "test", "group": "_nogroup", "pool": "data", "size": size}
			cmd, err := build(Request{Action: action, ResourceKey: "filesystem/cephfs/subvolume"}, params)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, arg := range cmd.args {
				if arg == size {
					found = true
				}
			}
			if size != "0" && !found {
				t.Fatalf("quota lost in %s: %v", action, cmd.args)
			}
		}
		for _, size := range []any{0, 1024, nil, "", "01", "-1", "1.5", "1e3", "9223372036854775808"} {
			_, err := build(Request{Action: action, ResourceKey: "filesystem/cephfs/subvolume"}, map[string]any{"name": "test", "group": "_nogroup", "pool": "data", "size": size})
			if err == nil {
				t.Fatalf("%s accepted invalid quota %v", action, size)
			}
		}
	}
}

func TestCephFSQuotaDecimalBounds(t *testing.T) {
	for _, size := range []string{"1", "9007199254740993", "9223372036854775807"} {
		for _, req := range []Request{{Action: "subvolume.update", ResourceKey: "filesystem/cephfs/subvolume/home"}, groupUpdateRequest(nil)} {
			cmd, err := build(req, map[string]any{"size": size})
			if err != nil {
				t.Fatal(err)
			}
			if cmd.args[5] != size {
				t.Fatalf("rounded size: %v", cmd.args)
			}
		}
	}
	for _, size := range []any{1024, "", "0", "-1", "+1", "01", "1e3", "1.5", "9223372036854775808"} {
		if _, err := cephFSQuotaSize(map[string]any{"size": size}); err == nil {
			t.Fatalf("invalid size accepted: %v", size)
		}
	}
}
