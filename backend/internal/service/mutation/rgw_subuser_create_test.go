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

func TestRGWSubuserCreation(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, kind := range []string{"s3", "swift"} {
		for _, permission := range []string{"none", "read", "write", "readwrite", "full"} {
			t.Run(kind+"/"+permission, func(t *testing.T) {
				params := map[string]any{"uid": "tenant$ns$user", "action": "create", "subuser": "sub", "confirm_subuser": "tenant$ns$user:sub", "subuser_permission": permission, "key_type": kind, "secret_key": "Private+/=secret"}
				if kind == "s3" {
					params["access_key"] = "ACCESS123"
				}
				key := map[string]any{"user": "tenant$ns$user:sub", "secret_key": params["secret_key"], "active": true}
				if kind == "s3" {
					key["access_key"] = params["access_key"]
				}
				info := map[string]any{"full_user_id": "tenant$ns$user", "subusers": []any{map[string]any{"id": "tenant$ns$user:sub", "permissions": rgwSubuserPermissions[permission]}}, "keys": []any{}, "swift_keys": []any{}}
				keyField := "keys"
				if kind == "swift" {
					keyField = "swift_keys"
				}
				info[keyField] = []any{key}
				post, _ := json.Marshal(info)
				before := `{"full_user_id":"tenant$ns$user","subusers":[],"keys":[],"swift_keys":[]}`
				for _, scenario := range []string{"success", "existing", "orphan-key", "pre-error", "write-error", "post-error", "missing-key", "wrong-secret", "inactive-key", "wrong-permission"} {
					runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.subuser.pre_check": before, "rgw_user.subuser.post_check": string(post), "rgw_user.subuser": string(post)}}
					wantCount := 3
					switch scenario {
					case "existing":
						runner.outputs["rgw_user.subuser.pre_check"] = string(post)
						wantCount = 1
					case "orphan-key":
						runner.outputs["rgw_user.subuser.pre_check"] = `{"full_user_id":"tenant$ns$user","subusers":[],"keys":[],"swift_keys":[{"user":"tenant$ns$user:sub"}]}`
						wantCount = 1
					case "pre-error":
						runner.failID = "rgw_user.subuser.pre_check"
						wantCount = 1
					case "write-error":
						runner.failID = "rgw_user.subuser"
						wantCount = 2
					case "post-error":
						runner.failID = "rgw_user.subuser.post_check"
					case "missing-key":
						runner.outputs["rgw_user.subuser.post_check"] = before
					case "wrong-secret":
						runner.outputs["rgw_user.subuser.post_check"] = strings.ReplaceAll(string(post), "Private+/=secret", "wrong-secret")
					case "inactive-key":
						runner.outputs["rgw_user.subuser.post_check"] = strings.ReplaceAll(string(post), `"active":true`, `"active":false`)
					case "wrong-permission":
						encodedPermission, _ := json.Marshal(rgwSubuserPermissions[permission])
						runner.outputs["rgw_user.subuser.post_check"] = strings.ReplaceAll(string(post), `"permissions":`+string(encodedPermission), `"permissions":"unexpected"`)
					}
					service.executor = runner
					result, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.subuser", ResourceKey: "rgw/user/tenant$ns$user", Parameters: params})
					if (err == nil) != (scenario == "success") || len(runner.specs) != wantCount {
						t.Fatalf("%s: %v %+v", scenario, err, runner.specs)
					}
					if wantCount > 1 && scenario != "success" {
						var actionErr *cephdomain.ActionError
						if !errors.As(err, &actionErr) || actionErr.Retryable {
							t.Fatalf("unsafe retry: %v", err)
						}
					}
					if wantCount > 1 {
						access := permission
						if permission == "none" {
							access = ""
						}
						want := []string{"subuser", "create", "--uid", "tenant$ns$user", "--subuser=sub", "--access=" + access, "--key-type=" + kind, "--secret-key=Private+/=secret"}
						masked := map[int]struct{}{7: {}}
						if kind == "s3" {
							want = append(want, "--access-key=ACCESS123")
							masked[8] = struct{}{}
						}
						want = append(want, "--format", "json")
						if !reflect.DeepEqual(runner.specs[1].Args, want) || !reflect.DeepEqual(runner.specs[1].SensitiveArgs, masked) {
							t.Fatalf("wrong command or masking: %+v", runner.specs[1])
						}
					}
					encoded, _ := json.Marshal(result)
					if strings.Contains(string(encoded), "Private") || strings.Contains(string(encoded), "ACCESS123") || (err != nil && strings.Contains(err.Error(), "Private")) {
						t.Fatal("credential leaked into result")
					}
				}
			})
		}
	}
}

func TestRGWSubuserCreationCredentialValidation(t *testing.T) {
	base := func() map[string]any {
		return map[string]any{"uid": "user", "action": "create", "subuser": "sub", "confirm_subuser": "user:sub", "subuser_permission": "read", "key_type": "s3", "access_key": "ACCESS123", "secret_key": "secret"}
	}
	for field, values := range map[string][]any{"key_type": {nil, "", " s3", "other"}, "secret_key": {nil, "", " secret", "secret ", "a\tb", "a\x00b", strings.Repeat("s", 257)}, "access_key": {nil, "", "bad/key", "--option", strings.Repeat("A", 129)}} {
		for _, value := range values {
			p := base()
			p[field] = value
			if _, err := build(Request{Action: "rgw_user.subuser"}, p); err == nil {
				t.Fatalf("accepted %s", field)
			}
		}
	}
	for _, action := range []string{"modify", "rm"} {
		p := base()
		p["action"] = action
		if _, err := build(Request{Action: "rgw_user.subuser"}, p); err == nil {
			t.Fatal("accepted credentials for existing subuser")
		}
	}
	p := base()
	p["key_type"] = "swift"
	if _, err := build(Request{Action: "rgw_user.subuser"}, p); err == nil {
		t.Fatal("accepted Swift access key")
	}
}
