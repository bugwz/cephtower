package handler

import (
	"net/http/httptest"
	"testing"
)

func TestCephFSReadKeysIncludeDefaultAndNamedGroups(t *testing.T) {
	for _, group := range []string{"", "_nogroup", "team-a", "team-b"} {
		body := map[string]any{"fs": "cephfs", "subvolume": "same", "group": group, "snap": "daily"}
		wantGroup := group
		if wantGroup == "" {
			wantGroup = "_nogroup"
		}
		if got := readResourceKey("subvolume", body); got != "cephfs/"+wantGroup+"/same" {
			t.Fatalf("subvolume key=%s", got)
		}
		if got := readResourceKey("cephfs_snapshot", body); got != "cephfs/"+wantGroup+"/same/daily" {
			t.Fatalf("snapshot key=%s", got)
		}
	}
}

func TestSnapshotParentFilterIsScopedOnlyWhenSubvolumeIsSelected(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	for _, body := range []map[string]any{{}, {"fs": "cephfs"}, {"group": "users"}} {
		filter := storeResourceFilter("cephfs_snapshot", 50, 0, r, body)
		if filter.ParentKind != "" || filter.ParentKey != "" {
			t.Fatalf("unexpected parent filter=%+v", filter)
		}
	}
	filter := storeResourceFilter("cephfs_snapshot", 50, 0, r, map[string]any{"fs": "cephfs", "subvolume": "same", "group": "users"})
	if filter.ParentKind != "subvolume" || filter.ParentKey != "cephfs/users/same" {
		t.Fatalf("filter=%+v", filter)
	}
}
