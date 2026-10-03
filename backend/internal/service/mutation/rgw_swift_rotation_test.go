package mutation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWSwiftRotation(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, active := range []bool{false, true} {
		params := map[string]any{"uid": "tenant$ns$user", "action": "rotate-swift-key", "subuser": "sub", "confirm_subuser": "tenant$ns$user:sub", "expected_key_active": active, "secret_key": "NEWsecret"}
		info := func(secret string) string {
			return `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:sub","permissions":"read"}],"swift_keys":[{"user":"tenant$ns$user:sub","secret_key":"` + secret + `","active":` + strconv.FormatBool(active) + `}]}`
		}
		for _, scenario := range []string{"success", "missing", "same-secret", "state-drift", "duplicate", "pre-error", "write-error", "post-error", "old-secret", "post-state-drift", "wrong-user"} {
			t.Run(strconv.FormatBool(active)+"/"+scenario, func(t *testing.T) {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.subuser.pre_check": info("OLDsecret"), "rgw_user.subuser.post_check": info("NEWsecret")}}
				count := 3
				switch scenario {
				case "missing":
					runner.outputs["rgw_user.subuser.pre_check"] = `{"full_user_id":"tenant$ns$user","subusers":[],"swift_keys":[]}`
					count = 1
				case "same-secret":
					runner.outputs["rgw_user.subuser.pre_check"] = info("NEWsecret")
					count = 1
				case "state-drift":
					runner.outputs["rgw_user.subuser.pre_check"] = strings.ReplaceAll(info("OLDsecret"), `"active":`+strconv.FormatBool(active), `"active":`+strconv.FormatBool(!active))
					count = 1
				case "duplicate":
					runner.outputs["rgw_user.subuser.pre_check"] = strings.ReplaceAll(info("OLDsecret"), `"swift_keys":[`, `"swift_keys":[{"user":"tenant$ns$user:sub","secret_key":"OLDsecret","active":`+strconv.FormatBool(active)+`},`)
					count = 1
				case "pre-error":
					runner.failID = "rgw_user.subuser.pre_check"
					count = 1
				case "write-error":
					runner.failID = "rgw_user.subuser"
					count = 2
				case "post-error":
					runner.failID = "rgw_user.subuser.post_check"
				case "old-secret":
					runner.outputs["rgw_user.subuser.post_check"] = info("OLDsecret")
				case "post-state-drift":
					runner.outputs["rgw_user.subuser.post_check"] = strings.ReplaceAll(info("NEWsecret"), `"active":`+strconv.FormatBool(active), `"active":`+strconv.FormatBool(!active))
				case "wrong-user":
					runner.outputs["rgw_user.subuser.post_check"] = strings.ReplaceAll(info("NEWsecret"), "tenant$ns$user", "other")
				}
				service.executor = runner
				result, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.subuser", ResourceKey: "rgw/user/tenant$ns$user", Parameters: params})
				if (err == nil) != (scenario == "success") || len(runner.specs) != count {
					t.Fatalf("%s: %v %+v", scenario, err, runner.specs)
				}
				if count > 1 && err != nil {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Retryable {
						t.Fatalf("unsafe retry: %v", err)
					}
				}
				for i, spec := range runner.specs {
					want := []string{"user", "info", "--uid", "tenant$ns$user", "--format", "json"}
					if i == 1 {
						want = []string{"key", "create", "--uid", "tenant$ns$user", "--subuser=sub", "--key-type=swift", "--key-active=" + strconv.FormatBool(active), "--secret-key=NEWsecret", "--format", "json"}
						if !reflect.DeepEqual(spec.SensitiveArgs, map[int]struct{}{7: {}}) {
							t.Fatal("secret argument is not masked")
						}
					}
					if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i == 1) || !reflect.DeepEqual(spec.Args, want) {
						t.Fatalf("wrong step %d: %+v", i, spec)
					}
				}
				if result.Details != nil {
					if _, ok := result.Details.(map[string]any)["native_output"]; ok {
						t.Fatal("native credential output returned")
					}
				}
			})
		}
	}
}

func TestRGWSwiftRotationValidation(t *testing.T) {
	for field, values := range map[string][]any{"expected_key_active": {nil, "false", 0}, "secret_key": {nil, "", " new", "new\n", strings.Repeat("s", 257)}, "subuser_permission": {"read"}, "key_type": {"swift", "s3"}, "access_key": {"ACCESS"}} {
		for _, value := range values {
			p := map[string]any{"uid": "user", "action": "rotate-swift-key", "subuser": "sub", "confirm_subuser": "user:sub", "expected_key_active": false, "secret_key": "new"}
			p[field] = value
			if _, err := build(Request{Action: "rgw_user.subuser"}, p); err == nil {
				t.Fatalf("accepted %s=%v", field, value)
			}
		}
	}
}
