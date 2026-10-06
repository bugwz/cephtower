package mutation

import (
	"context"
	"testing"
)

func TestOSDRemovalSchedulingConfirmation(t *testing.T) {
	const confirmed = "Scheduled OSD(s) for removal."
	const warning = "\nVG/LV for the OSDs won't be zapped (--zap wasn't passed).\nRun the `ceph-volume lvm zap` command with `--destroy` against the VG/LV if you want them to be destroyed."
	for _, tc := range []struct {
		output     string
		zap, valid bool
	}{
		{confirmed, true, true}, {confirmed + warning + "\n", false, true},
		{"Unable to find OSDs: ['0']", false, false}, {"", false, false},
		{confirmed, false, false}, {confirmed + warning, true, false},
		{confirmed + "\nerror", true, false},
	} {
		for _, preserve := range []bool{false, true} {
			service, _, id := newCephUserService(t)
			runner := &removalStopExecutor{output: tc.output, queue: "No OSD remove/replace operations reported"}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.delete", ResourceKey: "osd/0", Parameters: map[string]any{"zap": tc.zap, "preserve_id": preserve}})
			if (err == nil) != tc.valid {
				t.Fatalf("%+v preserve=%v err=%v", tc, preserve, err)
			}
			wantCalls := 1
			if tc.valid {
				wantCalls = 2
			}
			if len(runner.specs) != wantCalls {
				t.Fatal(runner.specs)
			}
		}
	}
}
