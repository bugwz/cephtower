package mutation

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWS3KeyRotation(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, sub := range []string{"", "sub"} {
		for _, active := range []bool{false, true} {
			owner := "tenant$ns$user"
			if sub != "" {
				owner += ":" + sub
			}
			p := map[string]any{"uid": "tenant$ns$user", "confirm_owner": owner, "access_key": "ACCESS123", "secret_key": "NEWsecret"}
			if sub != "" {
				p["subuser"] = sub
			}
			info := func(secret string) string {
				return `{"full_user_id":"tenant$ns$user","subusers":[{"id":"tenant$ns$user:sub"}],"keys":[{"user":"` + owner + `","access_key":"ACCESS123","secret_key":"` + secret + `","active":` + strconv.FormatBool(active) + `}]}`
			}
			for _, scenario := range []string{"success", "same-secret", "missing", "wrong-owner", "duplicate", "unknown-state", "pre-error", "write-error", "post-error", "old-secret", "state-changed"} {
				t.Run(sub+"/"+strconv.FormatBool(active)+"/"+scenario, func(t *testing.T) {
					runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_key.update.pre_check": info("OLDsecret"), "rgw_key.update.post_check": info("NEWsecret")}}
					count := 3
					switch scenario {
					case "same-secret":
						runner.outputs["rgw_key.update.pre_check"] = info("NEWsecret")
						count = 1
					case "missing":
						runner.outputs["rgw_key.update.pre_check"] = `{"full_user_id":"tenant$ns$user","keys":[]}`
						count = 1
					case "wrong-owner":
						runner.outputs["rgw_key.update.pre_check"] = strings.ReplaceAll(info("OLDsecret"), `"user":"`+owner+`"`, `"user":"other"`)
						count = 1
					case "duplicate":
						runner.outputs["rgw_key.update.pre_check"] = strings.ReplaceAll(info("OLDsecret"), `"keys":[`, `"keys":[{"user":"`+owner+`","access_key":"ACCESS123","secret_key":"OLDsecret","active":true},`)
						count = 1
					case "unknown-state":
						runner.outputs["rgw_key.update.pre_check"] = strings.ReplaceAll(info("OLDsecret"), `"active":`+strconv.FormatBool(active), `"active":null`)
						count = 1
					case "pre-error":
						runner.failID = "rgw_key.update.pre_check"
						count = 1
					case "write-error":
						runner.failID = "rgw_key.update"
						count = 2
					case "post-error":
						runner.failID = "rgw_key.update.post_check"
					case "old-secret":
						runner.outputs["rgw_key.update.post_check"] = info("OLDsecret")
					case "state-changed":
						runner.outputs["rgw_key.update.post_check"] = strings.ReplaceAll(info("NEWsecret"), `"active":`+strconv.FormatBool(active), `"active":`+strconv.FormatBool(!active))
					}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_key.update", ResourceKey: "rgw/user/tenant$ns$user/key", Parameters: p})
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
							want = []string{"key", "create", "--uid", "tenant$ns$user", "--key-type=s3", "--access-key=ACCESS123", "--secret-key=NEWsecret", "--format", "json"}
							if !reflect.DeepEqual(spec.SensitiveArgs, map[int]struct{}{5: {}, 6: {}}) {
								t.Fatal("unmasked credential")
							}
						}
						if spec.Mutating != (i == 1) || !reflect.DeepEqual(spec.Args, want) {
							t.Fatalf("unexpected step %d: %+v", i, spec)
						}
					}
				})
			}
		}
	}
}
