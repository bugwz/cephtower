package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWS3KeyDeletion(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, sub := range []string{"", "sub"} {
		owner := "tenant$ns$user"
		if sub != "" {
			owner += ":" + sub
		}
		p := map[string]any{"uid": "tenant$ns$user", "confirm_owner": owner, "access_key": "ACCESS123"}
		if sub != "" {
			p["subuser"] = sub
		}
		before := `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:sub"}],"keys":[{"user":"` + owner + `","access_key":"ACCESS123","active":false},{"user":"tenant$ns$user","access_key":"OTHERKEY"}]}`
		post := `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:sub"}],"keys":[{"user":"tenant$ns$user","access_key":"OTHERKEY"}]}`
		for _, scenario := range []string{"success", "wrong-owner", "missing", "duplicate", "pre-error", "write-error", "post-error", "retained", "unknown-list", "wrong-user"} {
			t.Run(sub+"/"+scenario, func(t *testing.T) {
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_key.delete.pre_check": before, "rgw_key.delete.post_check": post}}
				count := 3
				switch scenario {
				case "wrong-owner":
					runner.outputs["rgw_key.delete.pre_check"] = strings.ReplaceAll(before, `"user":"`+owner+`","access_key":"ACCESS123"`, `"user":"other:sub","access_key":"ACCESS123"`)
					count = 1
				case "missing":
					runner.outputs["rgw_key.delete.pre_check"] = post
					count = 1
				case "duplicate":
					runner.outputs["rgw_key.delete.pre_check"] = strings.ReplaceAll(before, `"keys":[`, `"keys":[{"user":"`+owner+`","access_key":"ACCESS123"},`)
					count = 1
				case "pre-error":
					runner.failID = "rgw_key.delete.pre_check"
					count = 1
				case "write-error":
					runner.failID = "rgw_key.delete"
					count = 2
				case "post-error":
					runner.failID = "rgw_key.delete.post_check"
				case "retained":
					runner.outputs["rgw_key.delete.post_check"] = before
				case "unknown-list":
					runner.outputs["rgw_key.delete.post_check"] = `{"full_user_id":"tenant$ns$user"}`
				case "wrong-user":
					runner.outputs["rgw_key.delete.post_check"] = strings.ReplaceAll(post, "tenant$ns$user", "other")
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_key.delete", ResourceKey: "rgw/user/tenant$ns$user/key", Parameters: p})
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
						want = []string{"key", "rm", "--uid", "tenant$ns$user", "--key-type=s3", "--access-key=ACCESS123", "--format", "json"}
						if !reflect.DeepEqual(spec.SensitiveArgs, map[int]struct{}{5: {}}) {
							t.Fatal("unmasked Access Key")
						}
					}
					if spec.Binary != executor.BinaryRGWAdmin || spec.Mutating != (i == 1) || !reflect.DeepEqual(spec.Args, want) {
						t.Fatalf("step %d: %+v", i, spec)
					}
				}
			})
		}
	}
}

func TestRGWS3DeleteValidation(t *testing.T) {
	for field, values := range map[string][]any{"uid": {nil, " user", "a/b", "-option"}, "subuser": {nil, "", "other:sub", "-sub"}, "confirm_owner": {nil, "other", "user "}, "access_key": {nil, "", "bad/key", " key"}, "secret_key": {"secret", nil}} {
		for _, value := range values {
			p := map[string]any{"uid": "user", "confirm_owner": "user", "access_key": "ACCESS"}
			p[field] = value
			if _, err := build(Request{Action: "rgw_key.delete"}, p); err == nil {
				t.Fatalf("accepted %s=%v", field, value)
			}
		}
	}
}
