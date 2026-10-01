package mutation

import "testing"

func TestSnapshotScheduleNativeIntervals(t *testing.T) {
	for _, action := range []string{"snapshot_schedule.create", "snapshot_schedule.action"} {
		for _, interval := range []string{"1m", "2h", "3d", "4w", "5M", "6y", "01d"} {
			cmd, err := build(Request{Action: action, ResourceKey: "filesystem/data/snapshot-schedule"}, map[string]any{"path": "/project", "schedule": interval, "start": "2026-10-01T00:00:00", "action": "activate"})
			if err != nil {
				t.Fatalf("%s %s: %v", action, interval, err)
			}
			want := interval
			if action == "snapshot_schedule.action" {
				want = "--repeat=" + interval
			}
			if cmd.args[4] != want {
				t.Fatalf("unexpected interval argument: %v", cmd.args)
			}
		}
		for _, interval := range []string{"0h", "00h", "-1d", "1.5h", "1Y", "1H", "1h2d", "1", "1 h"} {
			_, err := build(Request{Action: action, ResourceKey: "filesystem/data/snapshot-schedule"}, map[string]any{"path": "/project", "schedule": interval, "start": "2026-10-01T00:00:00", "action": "activate"})
			if err == nil {
				t.Fatalf("%s accepted invalid interval %q", action, interval)
			}
		}
	}
}

func TestSnapshotScheduleGroupRequiresSubvolume(t *testing.T) {
	for _, action := range []string{"snapshot_schedule.create", "snapshot_schedule.action"} {
		params := map[string]any{"path": "/", "schedule": "1h", "start": "2026-10-01T00:00:00", "action": "activate", "group": "team"}
		request := Request{Action: action, ResourceKey: "filesystem/data/snapshot-schedule"}
		if _, err := build(request, params); err == nil {
			t.Fatalf("%s accepted group without subvolume", action)
		}
		params["subvol"] = "home"
		if _, err := build(request, params); err != nil {
			t.Fatal(err)
		}
	}
}
