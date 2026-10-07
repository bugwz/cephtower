package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRBDPeerEditReadback(t *testing.T) {
	for field, key := range map[string]string{"site-name": "site_name", "client": "client_name", "direction": "direction", "mon-host": "mon_host"} {
		for _, scenario := range []string{"success", "mismatch", "missing", "duplicate", "pre_failure", "write_failure", "post_failure"} {
			t.Run(field+"/"+scenario, func(t *testing.T) {
				value := "updated"
				if field == "direction" {
					value = "rx-only"
				}
				peer := map[string]any{"uuid": "peer-a", key: value, "key": "SECRET-NOT-RETURNED"}
				if scenario == "mismatch" {
					peer[key] = "old"
				}
				if scenario == "missing" {
					delete(peer, key)
				}
				peers := []any{peer}
				if scenario == "duplicate" {
					peers = append(peers, peer)
				}
				raw, _ := json.Marshal(map[string]any{"mode": "image", "peers": peers})
				runner := &directoryRenameExecutor{outputs: map[string]string{"rbd_mirroring.peer.pre_check": `{"mode":"image","peers":[{"uuid":"peer-a"}]}`, "rbd_mirroring.peer.post_check": string(raw)}}
				code, calls := "post_check_failed", 3
				switch scenario {
				case "success":
					code = ""
				case "pre_failure":
					runner.failID = "rbd_mirroring.peer.pre_check"
					code = "invalid_request"
					calls = 1
				case "write_failure":
					runner.failID = "rbd_mirroring.peer"
					code = "ceph_command_failed"
					calls = 2
				case "post_failure":
					runner.failID = "rbd_mirroring.peer.post_check"
				}
				s, _, clusterID := newCephUserService(t)
				s.executor = runner
				result, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rbd_mirroring.peer", Parameters: map[string]any{"pool": "pool-a", "uuid": "peer-a", "action": "set", "field": field, "value": value}})
				if code == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var ae *cephdomain.ActionError
					if !errors.As(err, &ae) || ae.Code != code || ae.Retryable {
						t.Fatalf("err=%v want %s", err, code)
					}
				}
				encoded, _ := json.Marshal(result)
				if strings.Contains(string(encoded), "SECRET") || err != nil && strings.Contains(err.Error(), "SECRET") {
					t.Fatal("credential exposed")
				}
				if len(runner.specs) != calls {
					t.Fatalf("calls=%d", len(runner.specs))
				}
				for i, spec := range runner.specs {
					want := []string{"mirror", "pool", "info", "pool-a", "--format", "json"}
					if i == 1 {
						want = []string{"mirror", "pool", "peer", "set", "pool-a", "peer-a", field, value}
					} else if i == 2 && field == "mon-host" {
						want = append(want, "--all")
					}
					if !reflect.DeepEqual(spec.Args, want) || spec.Mutating != (i == 1) {
						t.Fatalf("command=%#v", spec)
					}
				}
			})
		}
	}
}
