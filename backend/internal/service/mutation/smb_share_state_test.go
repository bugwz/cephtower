package mutation

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestSMBShareLifecycleReadback(t *testing.T) {
	for _, action := range []string{"smb_share.delete"} {
		t.Run(action, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			request := Request{ClusterID: id, Action: action, ResourceKey: "smb/share/" + base64.RawURLEncoding.EncodeToString([]byte("a\x00docs")), Parameters: map[string]any{"cluster": "a", "name": "docs", "filesystem": "fs"}}
			good, stale := `["docs"]`, `[]`
			if action == "smb_share.delete" {
				good, stale = stale, good
				request.Parameters["name"] = "not-the-target"
			}
			runner := &directoryRenameExecutor{outputs: map[string]string{action + ".post_check": good}}
			service.executor = runner
			if _, err := service.Execute(context.Background(), request); err != nil {
				t.Fatal(err)
			}
			if len(runner.specs) != 2 || runner.specs[1].Mutating || runner.specs[1].Args[3] != "a" {
				t.Fatalf("incorrect cluster-scoped readback: %v", runner.specs)
			}
			for _, bad := range []string{stale, `null`, `{}`, `[1]`, `[""]`, `["docs","docs"]`, good + ` {}`} {
				runner.outputs[action+".post_check"] = bad
				_, err := service.Execute(context.Background(), request)
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Code != "post_check_failed" {
					t.Fatalf("invalid readback %s accepted: %v", bad, err)
				}
			}
		})
	}
}
