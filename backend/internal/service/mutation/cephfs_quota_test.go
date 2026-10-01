package mutation

import "testing"

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
