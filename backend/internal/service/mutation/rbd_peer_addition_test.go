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

func TestRBDPeerAdditionReadback(t *testing.T) {
	const id = "12345678-1234-1234-1234-123456789abc"
	for _, scenario := range []string{"success", "existing", "bad_uuid", "extra_line", "empty_uuid", "site_name", "client_name", "direction", "missing", "duplicate", "malformed_pre", "malformed_post", "pre_failure", "write_failure", "post_failure"} {
		t.Run(scenario, func(t *testing.T) {
			peer := map[string]any{"uuid": id, "site_name": "remote", "client_name": "client.mirror", "direction": "rx-tx", "key": "SECRET-NOT-RETURNED"}
			if scenario == "site_name" || scenario == "client_name" || scenario == "direction" {
				peer[scenario] = "different"
			}
			peers := []any{peer}
			if scenario == "missing" {
				peers = []any{}
			}
			if scenario == "duplicate" {
				peers = append(peers, map[string]any{"uuid": strings.ToUpper(id)})
			}
			raw, _ := json.Marshal(map[string]any{"mode": "image", "peers": peers})
			runner := &directoryRenameExecutor{outputs: map[string]string{"rbd_mirroring.peer.pre_check": `{"mode":"image","peers":[]}`, "rbd_mirroring.peer": id + "\n", "rbd_mirroring.peer.post_check": string(raw)}}
			code, calls := "post_check_failed", 3
			switch scenario {
			case "success":
				code = ""
			case "existing":
				runner.outputs["rbd_mirroring.peer.pre_check"] = `{"mode":"image","peers":[{"uuid":"` + strings.ToUpper(id) + `"}]}`
			case "bad_uuid":
				runner.outputs["rbd_mirroring.peer"] = "SECRET-NOT-RETURNED"
			case "extra_line":
				runner.outputs["rbd_mirroring.peer"] += "\n"
			case "empty_uuid":
				runner.outputs["rbd_mirroring.peer"] = ""
			case "malformed_pre":
				runner.outputs["rbd_mirroring.peer.pre_check"] = `{}`
				code, calls = "invalid_request", 1
			case "malformed_post":
				runner.outputs["rbd_mirroring.peer.post_check"] = `{}`
			case "pre_failure":
				runner.failID = "rbd_mirroring.peer.pre_check"
				code, calls = "invalid_request", 1
			case "write_failure":
				runner.failID = "rbd_mirroring.peer"
				code, calls = "ceph_command_failed", 2
			case "post_failure":
				runner.failID = "rbd_mirroring.peer.post_check"
			}
			s, _, clusterID := newCephUserService(t)
			s.executor = runner
			result, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rbd_mirroring.peer", Parameters: map[string]any{"pool": "pool-a", "action": "add", "remote_cluster": "remote", "remote_client": "client.mirror", "direction": "rx-tx"}})
			if code == "" {
				want := map[string]any{"pool": "pool-a", "peer_uuid": id, "created": true, "verified": true}
				if err != nil || !reflect.DeepEqual(result.Details, want) {
					t.Fatalf("result=%#v err=%v", result, err)
				}
			} else {
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Code != code || ae.Retryable {
					t.Fatalf("err=%v want %s", err, code)
				}
			}
			encoded, _ := json.Marshal(result)
			if strings.Contains(string(encoded), "SECRET") || err != nil && strings.Contains(err.Error(), "SECRET") {
				t.Fatal("raw output exposed")
			}
			if len(runner.specs) != calls {
				t.Fatalf("calls=%d want=%d", len(runner.specs), calls)
			}
			for i, spec := range runner.specs {
				want := []string{"mirror", "pool", "info", "pool-a", "--format", "json"}
				if i == 1 {
					want = []string{"mirror", "pool", "peer", "add", "pool-a", "--remote-cluster=remote", "--remote-client-name=client.mirror", "--direction=rx-tx"}
				}
				if !reflect.DeepEqual(spec.Args, want) || spec.Mutating != (i == 1) {
					t.Fatalf("command=%#v", spec)
				}
			}
		})
	}
}
