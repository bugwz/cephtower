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

func TestRGWUserCreationCredentials(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, withKey := range []bool{false, true} {
		p := map[string]any{"uid": "tenant$ns$user", "display_name": "User"}
		keys := `[]`
		if withKey {
			p["access_key"], p["secret_key"] = "ACCESS123", "SavedSecret"
			keys = `[{"user":"tenant$ns$user","access_key":"ACCESS123","secret_key":"SavedSecret","active":true}]`
		}
		valid := `{"full_user_id":"tenant$ns$user","display_name":"User","keys":` + keys + `,"swift_keys":[]}`
		for _, scenario := range []string{"success", "existing", "pre-error", "write-error", "post-error", "wrong-uid", "missing-keys", "wrong-keys", "swift-key"} {
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.create.pre_check": `[]`, "rgw_user.create.post_check": valid}}
			count := 3
			switch scenario {
			case "existing":
				runner.outputs["rgw_user.create.pre_check"] = `["tenant$ns$user"]`
				count = 1
			case "pre-error":
				runner.failID = "rgw_user.create.pre_check"
				count = 1
			case "write-error":
				runner.failID = "rgw_user.create"
				count = 2
			case "post-error":
				runner.failID = "rgw_user.create.post_check"
			case "wrong-uid":
				runner.outputs["rgw_user.create.post_check"] = strings.ReplaceAll(valid, "tenant$ns$user", "other")
			case "missing-keys":
				runner.outputs["rgw_user.create.post_check"] = `{"full_user_id":"tenant$ns$user"}`
			case "wrong-keys":
				runner.outputs["rgw_user.create.post_check"] = `{"full_user_id":"tenant$ns$user","keys":[{"user":"other"}],"swift_keys":[]}`
			case "swift-key":
				runner.outputs["rgw_user.create.post_check"] = strings.Replace(valid, `"swift_keys":[]`, `"swift_keys":[{"user":"tenant$ns$user:sub"}]`, 1)
			}
			service.executor = runner
			result, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.create", ResourceKey: "rgw/user/tenant$ns$user", Parameters: p})
			if (err == nil) != (scenario == "success") || len(runner.specs) != count {
				t.Fatalf("key=%v %s: %v %+v", withKey, scenario, err, runner.specs)
			}
			if err != nil && count > 1 {
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Retryable {
					t.Fatalf("unsafe retry: %v", err)
				}
			}
			encoded, _ := json.Marshal(result)
			if strings.Contains(string(encoded), "SavedSecret") || strings.Contains(string(encoded), "ACCESS123") {
				t.Fatal("credential in result")
			}
			if count > 1 {
				write := runner.specs[1]
				want := []string{"user", "create", "--uid", "tenant$ns$user", "--display-name", "User", "--generate-key=false", "--format", "json"}
				if withKey {
					want = []string{"user", "create", "--uid", "tenant$ns$user", "--display-name", "User", "--key-type=s3", "--access-key=ACCESS123", "--secret-key=SavedSecret", "--format", "json"}
					if !reflect.DeepEqual(write.SensitiveArgs, map[int]struct{}{7: {}, 8: {}}) {
						t.Fatal("unmasked credentials")
					}
				}
				if !reflect.DeepEqual(write.Args, want) || !write.Mutating {
					t.Fatalf("bad command: %+v", write)
				}
			}
		}
		if withKey {
			for _, bad := range []string{strings.Replace(valid, "SavedSecret", "OtherSecret", 1), strings.Replace(valid, `"active":true`, `"active":false`, 1), strings.Replace(valid, `"active":true`, `"active":null`, 1), strings.Replace(valid, "ACCESS123", "OtherAccess", 1)} {
				if rgwUserCreateKeysMatch([]byte(bad), p) {
					t.Fatalf("accepted mismatched key: %s", bad)
				}
			}
		}
	}
	for _, fields := range []map[string]any{{"access_key": "A"}, {"secret_key": "S"}, {"access_key": "A", "secret_key": ""}, {"access_key": "bad/key", "secret_key": "S"}, {"access_key": false, "secret_key": "S"}, {"access_key": "A", "secret_key": " padded"}} {
		fields["uid"], fields["display_name"] = "u", "User"
		if _, err := build(Request{Action: "rgw_user.create"}, fields); err == nil {
			t.Fatalf("accepted invalid credentials: %v", fields)
		}
	}
}
