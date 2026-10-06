package mutation

import "strings"

func osdRemovalScheduled(parameters map[string]any, raw []byte) bool {
	const confirmation = "Scheduled OSD(s) for removal."
	const warning = "\nVG/LV for the OSDs won't be zapped (--zap wasn't passed).\nRun the `ceph-volume lvm zap` command with `--destroy` against the VG/LV if you want them to be destroyed."
	expected := confirmation
	if zap, _ := parameters["zap"].(bool); !zap {
		expected += warning
	}
	return strings.TrimSpace(string(raw)) == expected
}
