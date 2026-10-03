package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func capsInfo(caps map[string]string) string {
	rows := []map[string]string{}
	for kind, permission := range caps {
		rows = append(rows, map[string]string{"type": kind, "perm": permission})
	}
	raw, _ := json.Marshal(map[string]any{"full_user_id": "tenant$ns$user", "caps": rows})
	return string(raw)
}

func TestRGWUserCapsBitSemantics(t *testing.T) {
	native := map[uint8]string{0: "<none>", 1: "read", 2: "write", 3: "*"}
	for _, action := range []string{"add", "rm"} {
		for permission, mask := range map[string]uint8{"read": 1, "write": 2, "read,write": 3, "*": 3} {
			for _, present := range []bool{false, true} {
				for bits := uint8(0); bits <= 3; bits++ {
					before := map[string]string{"usage": "read"}
					old := uint8(0)
					if present {
						before["users"] = native[bits]
						old = bits
					}
					p := map[string]any{"uid": "tenant$ns$user", "action": action, "type": "users", "permission": permission}
					expected, ok := rgwExpectedCaps([]byte(capsInfo(before)), p)
					want := old | mask
					if action == "rm" {
						want = old &^ mask
					}
					after := map[string]string{"usage": "read"}
					if action == "add" || want != 0 {
						after["users"] = native[want]
					}
					if !ok || !rgwUserCapsMatch([]byte(capsInfo(after)), "tenant$ns$user", expected) {
						t.Fatalf("%s %s present=%v bits=%d: %+v", action, permission, present, bits, expected)
					}
				}
			}
		}
	}
	for _, raw := range []string{`{}`, `null`, `{"full_user_id":"tenant$ns$user","caps":null}`, `{"full_user_id":"other","caps":[]}`, `{"full_user_id":"tenant$ns$user","caps":[{"type":"users","perm":"read, write"}]}`, `{"full_user_id":"tenant$ns$user","caps":[{"type":"users","perm":"read"},{"type":"users","perm":"write"}]}`, `{"full_user_id":"tenant$ns$user","caps":[{"perm":"*"}]}`} {
		if _, ok := rgwUserCaps([]byte(raw), "tenant$ns$user"); ok {
			t.Fatalf("accepted invalid snapshot: %s", raw)
		}
	}
}

func TestRGWUserCapsExecution(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, action := range []string{"add", "rm"} {
		before := capsInfo(map[string]string{"users": "read", "usage": "write"})
		after := capsInfo(map[string]string{"users": "*", "usage": "write"})
		permission := "write"
		if action == "rm" {
			permission = "read"
			after = capsInfo(map[string]string{"usage": "write"})
		}
		for _, scenario := range []string{"success", "pre-error", "invalid-pre", "write-error", "post-error", "unchanged", "unrelated-change", "wrong-user", "missing-caps"} {
			t.Run(action+"/"+scenario, func(t *testing.T) {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.caps.pre_check": before, "rgw_user.caps.post_check": after}}
				count := 3
				switch scenario {
				case "pre-error":
					runner.failID = "rgw_user.caps.pre_check"
					count = 1
				case "invalid-pre":
					runner.outputs["rgw_user.caps.pre_check"] = `{}`
					count = 1
				case "write-error":
					runner.failID = "rgw_user.caps"
					count = 2
				case "post-error":
					runner.failID = "rgw_user.caps.post_check"
				case "unchanged":
					runner.outputs["rgw_user.caps.post_check"] = before
				case "unrelated-change":
					runner.outputs["rgw_user.caps.post_check"] = capsInfo(map[string]string{"users": "*"})
				case "wrong-user":
					runner.outputs["rgw_user.caps.post_check"] = `{"full_user_id":"other","caps":[]}`
				case "missing-caps":
					runner.outputs["rgw_user.caps.post_check"] = `{"full_user_id":"tenant$ns$user"}`
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.caps", ResourceKey: "rgw/user/tenant$ns$user", Parameters: map[string]any{"uid": "tenant$ns$user", "action": action, "type": "users", "permission": permission}})
				if (err == nil) != (scenario == "success") || len(runner.specs) != count {
					t.Fatalf("%s: %v %+v", scenario, err, runner.specs)
				}
				if err != nil && count > 1 {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Retryable || (count == 3 && actionErr.Code != "post_check_failed") {
						t.Fatalf("unsafe failure: %v", err)
					}
				}
				for index, spec := range runner.specs {
					want := []string{"user", "info", "--uid", "tenant$ns$user", "--format", "json"}
					if index == 1 {
						want = []string{"caps", action, "--uid", "tenant$ns$user", "--caps", "users=" + permission, "--format", "json"}
					}
					if spec.Mutating != (index == 1) || !reflect.DeepEqual(spec.Args, want) {
						t.Fatalf("unexpected step: %+v", spec)
					}
				}
			})
		}
	}
}
