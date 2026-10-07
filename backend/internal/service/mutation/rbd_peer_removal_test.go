package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRBDPeerRemoval(t *testing.T) {
	const before = `{"mode":"image","peers":[{"uuid":"peer-a"},{"uuid":"peer-b"}]}`
	const after = `{"mode":"image","peers":[{"uuid":"peer-b"}]}`
	for _, tc := range []struct {
		pre, post, fail, code string
		calls                 int
	}{
		{before, after, "", "", 3},
		{before, before, "", "post_check_failed", 3},
		{before, `null`, "", "post_check_failed", 3},
		{before, `{"mode":"image"}`, "", "post_check_failed", 3},
		{after, after, "", "invalid_request", 1},
		{`{"mode":"image","peers":[{"uuid":"peer-a"},{"uuid":"PEER-A"}]}`, after, "", "invalid_request", 1},
		{before, after, "rbd_mirroring.peer.pre_check", "invalid_request", 1},
		{before, after, "rbd_mirroring.peer", "ceph_command_failed", 2},
		{before, after, "rbd_mirroring.peer.post_check", "post_check_failed", 3},
	} {
		s, _, clusterID := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"rbd_mirroring.peer.pre_check": tc.pre, "rbd_mirroring.peer.post_check": tc.post}, failID: tc.fail}
		s.executor = runner
		result, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rbd_mirroring.peer", Parameters: map[string]any{"pool": "pool-a", "uuid": "peer-a", "action": "remove"}})
		if tc.code == "" {
			if err != nil || result.Details.(map[string]any)["verified"] != true {
				t.Fatalf("result=%#v err=%v", result, err)
			}
		} else {
			var actionErr *cephdomain.ActionError
			if !errors.As(err, &actionErr) || actionErr.Code != tc.code || actionErr.Retryable {
				t.Fatalf("want %s: %v", tc.code, err)
			}
		}
		if len(runner.specs) != tc.calls {
			t.Fatalf("commands=%#v", runner.specs)
		}
		for _, spec := range runner.specs {
			want := []string{"mirror", "pool", "info", "pool-a", "--format", "json"}
			if spec.Mutating {
				want = []string{"mirror", "pool", "peer", "remove", "pool-a", "peer-a"}
			}
			if !reflect.DeepEqual(spec.Args, want) {
				t.Fatalf("args=%v", spec.Args)
			}
		}
	}
}

func TestRBDPeerPresenceRejectsMalformedInventory(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"mode":"image","peers":null}`, `{"mode":"unknown","peers":[]}`, `{"mode":"image","peers":[null]}`, `{"mode":"image","peers":[{}]}`, `{"mode":"image","peers":[{"uuid":1}]}`, `{"mode":"image","peers":[{"uuid":" peer"}]}`, `{"mode":"image","peers":[{"uuid":"other"},{"uuid":"other"}]}`} {
		if rbdPeerPresence([]byte(raw), "peer-a", false) || rbdPeerPresence([]byte(raw), "peer-a", true) {
			t.Fatalf("accepted malformed inventory: %s", raw)
		}
	}
	if !rbdPeerPresence([]byte(`{"mode":"image","peers":[]}`), "peer-a", false) {
		t.Fatal("valid empty peer list did not confirm absence")
	}
}
