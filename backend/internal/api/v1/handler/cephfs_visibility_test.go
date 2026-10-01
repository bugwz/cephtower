package handler

import "testing"

func TestSubvolumeVisibilityGroupIdentity(t *testing.T) {
	for _, group := range []string{"users", "_nogroup", ""} {
		body := map[string]any{"fs": "subvolume", "subvolume": "group", "group": group}
		if group == "" {
			group = "_nogroup"
		}
		key := resourceKey("subvolume", "subvolume.snapshot_visibility", nil, body)
		if want := "filesystem/subvolume/subvolume/group/group/" + group; key != want {
			t.Fatalf("key=%s want=%s", key, want)
		}
		if got := resourceLookupKey("subvolume", key); got != "subvolume/"+group+"/group" {
			t.Fatalf("lookup=%s", got)
		}
	}
}
