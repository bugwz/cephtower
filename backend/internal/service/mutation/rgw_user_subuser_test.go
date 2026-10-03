package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWSubuserExecution(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	const before = `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:swift","permissions":"read"}]}`
	for _, subuser_permission := range []string{"none", "read", "write", "readwrite", "full", "remove"} {
		t.Run(subuser_permission, func(t *testing.T) {
			params := map[string]any{"uid": "tenant$ns$user", "subuser": "swift", "action": "modify", "subuser_permission": subuser_permission, "confirm_subuser": "tenant$ns$user:swift"}
			post := `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:swift","permissions":"` + rgwSubuserPermissions[subuser_permission] + `"}]}`
			wantWrite := []string{"subuser", "modify", "--uid", "tenant$ns$user", "--subuser=swift"}
			access := subuser_permission
			if subuser_permission == "none" {
				access = ""
			}
			if subuser_permission == "remove" {
				params["action"] = "rm"
				delete(params, "subuser_permission")
				post = `{"full_user_id":"tenant$ns$user","subusers":[],"keys":[{"user":"tenant$ns$user"}],"swift_keys":[]}`
				wantWrite[1] = "rm"
			} else {
				wantWrite = append(wantWrite, "--access="+access)
			}
			wantWrite = append(wantWrite, "--format", "json")
			for _, scenario := range []string{"success", "pre-invalid", "post-invalid", "pre-error", "write-error", "post-error"} {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.subuser.pre_check": before, "rgw_user.subuser.post_check": post}}
				wantCount := 3
				switch scenario {
				case "pre-invalid":
					runner.outputs["rgw_user.subuser.pre_check"] = `{}`
					wantCount = 1
				case "post-invalid":
					runner.outputs["rgw_user.subuser.post_check"] = `{}`
				case "pre-error":
					runner.failID = "rgw_user.subuser.pre_check"
					wantCount = 1
				case "write-error":
					runner.failID = "rgw_user.subuser"
					wantCount = 2
				case "post-error":
					runner.failID = "rgw_user.subuser.post_check"
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.subuser", ResourceKey: "rgw/user/tenant$ns$user", Parameters: params})
				if (err == nil) != (scenario == "success") {
					t.Fatalf("%s: %v", scenario, err)
				}
				if scenario == "write-error" || scenario == "post-invalid" || scenario == "post-error" {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Retryable {
						t.Fatalf("unsafe retry: %v", err)
					}
				}
				if len(runner.specs) != wantCount {
					t.Fatalf("%s: %+v", scenario, runner.specs)
				}
				for i, spec := range runner.specs {
					want := []string{"user", "info", "--uid", "tenant$ns$user", "--format", "json"}
					if i == 1 {
						want = wantWrite
					}
					if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i == 1) || !reflect.DeepEqual(spec.Args, want) {
						t.Fatalf("%s step %d: %+v", scenario, i, spec)
					}
				}
			}
		})
	}
}

func TestRGWSubuserValidation(t *testing.T) {
	for field, values := range map[string][]any{
		"uid":                {"", "wrong/user", "-user", " user", nil},
		"subuser":            {"", "other:swift", "user:swift", "--option", "a/b", "x\n", nil},
		"action":             {"create", "delete", " modify", nil},
		"subuser_permission": {"", "full-control", "read-write", true, nil},
		"confirm_subuser":    {"other:swift", "user:swift ", nil},
	} {
		for _, value := range values {
			p := map[string]any{"uid": "user", "subuser": "swift", "action": "modify", "subuser_permission": "read", "confirm_subuser": "user:swift"}
			p[field] = value
			if _, err := build(Request{Action: "rgw_user.subuser"}, p); err == nil {
				t.Fatalf("accepted %s = %#v", field, value)
			}
		}
	}
	p := map[string]any{"uid": "user", "subuser": "swift", "action": "rm", "subuser_permission": "read", "confirm_subuser": "user:swift"}
	if _, err := build(Request{Action: "rgw_user.subuser"}, p); err == nil {
		t.Fatal("accepted subuser_permission on removal")
	}
	delete(p, "subuser_permission")
	for _, raw := range []string{`{}`, `null`, `{`, `{"full_user_id":"other","subusers":[],"keys":[],"swift_keys":[]}`, `{"full_user_id":"user","subusers":[],"keys":[]}`, `{"full_user_id":"user","subusers":null,"keys":[],"swift_keys":[]}`, `{"full_user_id":"user","subusers":[],"keys":[{"user":"user:swift"}],"swift_keys":[]}`, `{"full_user_id":"user","subusers":[],"keys":[],"swift_keys":[{"user":"user:swift"}]}`, `{"full_user_id":"user","subusers":[],"keys":[{}],"swift_keys":[]}`, `{"full_user_id":"user","subusers":[{"id":"user:swift"}],"keys":[],"swift_keys":[]}`} {
		if rgwSubuserMatches([]byte(raw), p, false) {
			t.Fatalf("accepted incomplete removal: %s", raw)
		}
	}
	for _, raw := range []string{`{"full_user_id":"user","subusers":[{"id":"other:swift"}]}`, `{"full_user_id":"user","subusers":[{"id":"user:swift"},{"id":"user:swift"}]}`} {
		if rgwSubuserMatches([]byte(raw), p, true) {
			t.Fatalf("accepted invalid identity: %s", raw)
		}
	}
}
