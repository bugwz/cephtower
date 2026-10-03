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

func TestRGWS3KeyCreation(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, sub := range []string{"", "sub"} {
		owner := "tenant$ns$user"
		if sub != "" {
			owner += ":" + sub
		}
		p := map[string]any{"uid": "tenant$ns$user", "confirm_owner": owner, "access_key": "ACCESS123", "secret_key": "Private+/=secret"}
		if sub != "" {
			p["subuser"] = sub
		}
		before := `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:sub"}],"keys":[]}`
		post := `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:sub"}],"keys":[{"user":"` + owner + `","access_key":"ACCESS123","secret_key":"Private+/=secret","active":true}]}`
		for _, scenario := range []string{"success", "existing", "wrong-user", "pre-error", "write-error", "post-error", "wrong-owner", "wrong-secret", "inactive", "missing"} {
			t.Run(sub+"/"+scenario, func(t *testing.T) {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_key.create.pre_check": before, "rgw_key.create.post_check": post, "rgw_key.create": post}}
				count := 3
				switch scenario {
				case "existing":
					runner.outputs["rgw_key.create.pre_check"] = post
					count = 1
				case "wrong-user":
					runner.outputs["rgw_key.create.pre_check"] = strings.ReplaceAll(before, "tenant$ns$user", "other")
					count = 1
				case "pre-error":
					runner.failID = "rgw_key.create.pre_check"
					count = 1
				case "write-error":
					runner.failID = "rgw_key.create"
					count = 2
				case "post-error":
					runner.failID = "rgw_key.create.post_check"
				case "wrong-owner":
					runner.outputs["rgw_key.create.post_check"] = strings.ReplaceAll(post, `"user":"`+owner+`"`, `"user":"other"`)
				case "wrong-secret":
					runner.outputs["rgw_key.create.post_check"] = strings.ReplaceAll(post, "Private+/=secret", "different")
				case "inactive":
					runner.outputs["rgw_key.create.post_check"] = strings.ReplaceAll(post, `"active":true`, `"active":false`)
				case "missing":
					runner.outputs["rgw_key.create.post_check"] = before
				}
				service.executor = runner
				result, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_key.create", ResourceKey: "rgw/user/tenant$ns$user/key", Parameters: p})
				if (err == nil) != (scenario == "success") || len(runner.specs) != count {
					t.Fatalf("%s: %v %+v", scenario, err, runner.specs)
				}
				if err != nil && count > 1 {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Retryable {
						t.Fatalf("unsafe retry: %v", err)
					}
				}
				for i, spec := range runner.specs {
					want := []string{"user", "info", "--uid", "tenant$ns$user", "--format", "json"}
					if i == 1 {
						want = []string{"key", "create", "--uid", "tenant$ns$user", "--key-type=s3"}
						if sub != "" {
							want = append(want, "--subuser=sub")
						}
						start := len(want)
						want = append(want, "--access-key=ACCESS123", "--secret-key=Private+/=secret", "--format", "json")
						if !reflect.DeepEqual(spec.SensitiveArgs, map[int]struct{}{start: {}, start + 1: {}}) {
							t.Fatal("unmasked credential")
						}
					}
					if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i == 1) || !reflect.DeepEqual(spec.Args, want) {
						t.Fatalf("step %d: %+v", i, spec)
					}
				}
				encoded, _ := json.Marshal(result)
				if strings.Contains(string(encoded), "Private") || strings.Contains(string(encoded), "ACCESS123") {
					t.Fatal("credential output exposed")
				}
			})
		}
	}
}

func TestRGWS3KeyCreateValidation(t *testing.T) {
	for field, values := range map[string][]any{"uid": {nil, " user", "a/b", "-option"}, "subuser": {nil, "", "other:sub", "-sub"}, "confirm_owner": {nil, "other", "user "}, "access_key": {nil, "", "bad/key"}, "secret_key": {nil, "", "secret ", "a\nb"}} {
		for _, value := range values {
			p := map[string]any{"uid": "user", "confirm_owner": "user", "access_key": "ACCESS", "secret_key": "secret"}
			p[field] = value
			if _, err := build(Request{Action: "rgw_key.create"}, p); err == nil {
				t.Fatalf("accepted %s=%v", field, value)
			}
		}
	}
	p := map[string]any{"uid": "user", "subuser": "sub", "confirm_owner": "user:sub", "access_key": "ACCESS", "secret_key": "secret"}
	for _, raw := range []string{`{}`, `null`, `{"full_user_id":"user","keys":[]}`, `{"full_user_id":"user","subusers":[{"id":"user:sub"},{"id":"user:sub"}],"keys":[]}`, `{"full_user_id":"user","subusers":[{"id":"user:sub"}],"keys":[{"user":"user:other","access_key":"ACCESS"}]}`} {
		if rgwS3KeyCreationMatches([]byte(raw), p, true) {
			t.Fatalf("accepted uncertain absence or owner: %s", raw)
		}
	}
}
