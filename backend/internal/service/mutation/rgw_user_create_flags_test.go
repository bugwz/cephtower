package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWUserCreateFlags(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, system := range []bool{false, true} {
		for _, suspended := range []bool{false, true} {
			p := map[string]any{"uid": "tenant$ns$user", "display_name": "User", "system": system, "suspended": suspended}
			for _, scenario := range []string{"success", "write-error", "suspend-error", "wrong-system", "missing-system", "wrong-suspended", "missing-suspended", "invalid-suspended"} {
				if scenario == "suspend-error" && !suspended {
					continue
				}
				state := 0
				if suspended {
					state = 1
				}
				info := map[string]any{"full_user_id": "tenant$ns$user", "display_name": "User", "system": system, "suspended": state, "keys": []any{}, "swift_keys": []any{}}
				count := 3
				if suspended {
					count = 4
				}
				fail := ""
				switch scenario {
				case "write-error":
					fail = "rgw_user.create"
					count = 2
				case "suspend-error":
					fail = "rgw_user.create.step2"
					count = 3
				case "wrong-system":
					info["system"] = !system
				case "missing-system":
					delete(info, "system")
				case "wrong-suspended":
					info["suspended"] = 1 - state
				case "missing-suspended":
					delete(info, "suspended")
				case "invalid-suspended":
					info["suspended"] = true
				}
				raw, _ := json.Marshal(info)
				runner := &directoryRenameExecutor{failID: fail, outputs: map[string]string{"rgw_user.create.pre_check": "[]", "rgw_user.create.post_check": string(raw)}}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.create", ResourceKey: "rgw/user/tenant$ns$user", Parameters: p})
				if (err == nil) != (scenario == "success") || len(runner.specs) != count {
					t.Fatalf("system=%v suspended=%v %s: %v", system, suspended, scenario, err)
				}
				if err != nil {
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Retryable {
						t.Fatalf("unsafe failure: %v", err)
					}
				}
				for i, spec := range runner.specs {
					want := []string{"user", "info", "--uid", "tenant$ns$user", "--format", "json"}
					mutating := false
					if i == 0 {
						want = []string{"user", "list", "--format", "json"}
					}
					if i == 1 {
						want = []string{"user", "create", "--uid", "tenant$ns$user", "--system=" + strconv.FormatBool(system), "--display-name", "User", "--generate-key=false", "--format", "json"}
						mutating = true
					}
					if i == 2 && suspended {
						want = []string{"user", "suspend", "--uid", "tenant$ns$user", "--format", "json"}
						mutating = true
					}
					if !reflect.DeepEqual(spec.Args, want) || spec.Mutating != mutating {
						t.Fatalf("unexpected step: %+v", spec)
					}
				}
			}
		}
	}
	for _, field := range []string{"system", "suspended"} {
		for _, value := range []any{nil, "false", 0} {
			if _, err := build(Request{Action: "rgw_user.create"}, map[string]any{"uid": "u", "display_name": "User", field: value}); err == nil {
				t.Fatalf("accepted %s=%v", field, value)
			}
		}
	}
	p := map[string]any{"uid": "u", "display_name": "User", "system": true, "suspended": true, "access_key": "ACCESS123", "secret_key": "SavedSecret"}
	runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.create.pre_check": "[]", "rgw_user.create.post_check": `{"full_user_id":"u","display_name":"User","system":true,"suspended":1,"keys":[{"user":"u","access_key":"ACCESS123","secret_key":"SavedSecret","active":true}],"swift_keys":[]}`}}
	service.executor = runner
	if _, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.create", ResourceKey: "rgw/user/u", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 4 || !reflect.DeepEqual(runner.specs[1].SensitiveArgs, map[int]struct{}{8: {}, 9: {}}) || len(runner.specs[2].SensitiveArgs) != 0 {
		t.Fatal("suspension lost credential protection")
	}
}
