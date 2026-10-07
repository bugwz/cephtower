package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRBDPoolModeReadback(t *testing.T) {
	for _, mode := range []string{"disabled", "image", "pool"} {
		for _, scenario := range []string{"success", "mismatch", "missing", "null", "wrong_type", "trailing", "write_failure", "read_failure"} {
			t.Run(mode+"/"+scenario, func(t *testing.T) {
				raw := `{"mode":"` + mode + `","key":"SECRET-NOT-RETURNED"}`
				switch scenario {
				case "mismatch":
					raw = `{"mode":"init-only"}`
				case "missing":
					raw = `{}`
				case "null":
					raw = `null`
				case "wrong_type":
					raw = `{"mode":true}`
				case "trailing":
					raw += `{}`
				}
				runner := &directoryRenameExecutor{outputs: map[string]string{"rbd_mirroring.update.post_check": raw}}
				code, calls := "post_check_failed", 2
				switch scenario {
				case "success":
					code = ""
				case "write_failure":
					runner.failID = "rbd_mirroring.update"
					code, calls = "ceph_command_failed", 1
				case "read_failure":
					runner.failID = "rbd_mirroring.update.post_check"
				}
				s, _, clusterID := newCephUserService(t)
				s.executor = runner
				result, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rbd_mirroring.update", Parameters: map[string]any{"pool": "pool-a", "mode": mode}})
				if code == "" {
					if err != nil || !reflect.DeepEqual(result.Details, map[string]any{"pool": "pool-a", "mode": mode, "verified": true}) {
						t.Fatalf("result=%#v err=%v", result, err)
					}
				} else {
					var ae *cephdomain.ActionError
					if !errors.As(err, &ae) || ae.Code != code || ae.Retryable {
						t.Fatalf("err=%v want=%s", err, code)
					}
				}
				encoded, _ := json.Marshal(result)
				if strings.Contains(string(encoded), "SECRET") {
					t.Fatal("raw output exposed")
				}
				if len(runner.specs) != calls {
					t.Fatalf("calls=%d", len(runner.specs))
				}
				for i, spec := range runner.specs {
					want := []string{"mirror", "pool", "info", "pool-a", "--format", "json"}
					if i == 0 {
						want = []string{"mirror", "pool", "enable", "pool-a", mode}
						if mode == "disabled" {
							want = []string{"mirror", "pool", "disable", "pool-a"}
						}
					}
					if spec.Binary != executor.BinaryRBD || !reflect.DeepEqual(spec.Args, want) || spec.Mutating != (i == 0) {
						t.Fatalf("command=%#v", spec)
					}
				}
			})
		}
	}
}
