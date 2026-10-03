package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupSyncPipeUpdate(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, mode := range []string{"system", "user"} {
			for _, scenario := range []string{"success", "priority", "missing", "duplicate", "stale", "unchanged", "write", "post_check", "period.commit", "published_policy_check"} {
				if realm == "" && (scenario == "period.commit" || scenario == "published_policy_check") {
					continue
				}
				t.Run(realm+"/"+mode+"/"+scenario, func(t *testing.T) {
					service, _, cluster := newCephUserService(t)
					encode := func(v any) string {
						b, e := json.Marshal(v)
						if e != nil {
							t.Fatal(e)
						}
						return string(b)
					}
					source := map[string]any{"bucket": "old/photos:instance", "zones": []any{"orphan"}}
					dest := map[string]any{"bucket": "*", "zones": []any{"*"}}
					nativeParams := map[string]any{"mode": "user", "user": "old$u", "priority": json.Number("9007199254740993"), "source": map[string]any{"filter": map[string]any{"prefix": "photos/", "tags": []any{map[string]any{"key": "k", "value": "v"}}}}, "dest": map[string]any{"storage_class": "COLD", "acl_translation": map[string]any{"owner": "owner"}}}
					pipe := map[string]any{"id": "p", "source": source, "dest": dest, "params": nativeParams}
					other := map[string]any{"id": "other", "params": map[string]any{"priority": json.Number("9007199254740993")}}
					group := map[string]any{"id": "g", "status": "allowed", "data_flow": map[string]any{"symmetrical": []any{map[string]any{"id": "f", "zones": []any{"orphan"}}}}, "pipes": []any{pipe, other}}
					if scenario == "missing" {
						group["pipes"] = []any{other}
					}
					if scenario == "duplicate" {
						group["pipes"] = []any{pipe, pipe, other}
					}
					params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "pipe_id": "p", "expected_group": encode(group), "source_tenant": "team", "source_bucket": "new", "source_bucket_id": "id", "dest_bucket": "archive", "mode": mode}
					if mode == "user" {
						params["user"] = "new$u"
					}
					if scenario == "unchanged" {
						params["source_tenant"] = "old"
						params["source_bucket"] = "photos"
						params["source_bucket_id"] = "instance"
						params["dest_bucket"] = "*"
						params["mode"] = "user"
						params["user"] = "old$u"
					}
					zg := map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{}, "sync_policy": map[string]any{"groups": []any{group}}}
					before := encode(zg)
					if scenario == "priority" {
						params["priority"] = -2147483648
						nativeParams["priority"] = json.Number("-2147483648")
					}
					source["bucket"] = "team/new:id"
					dest["bucket"] = "archive"
					nativeParams["mode"] = mode
					if mode == "user" {
						nativeParams["user"] = "new$u"
					}
					after := encode(zg)
					period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{zg}}})
					runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": before, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "post_check": after, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
					count := 3
					if realm != "" {
						count = 9
					}
					switch scenario {
					case "missing", "duplicate", "unchanged":
						count = 1
					case "stale":
						params["expected_group"] = encode(group)
						count = 1
					case "write":
						runner.failure = "write"
						count = 2
						if realm != "" {
							count++
						}
					case "post_check":
						runner.bodies["post_check"] = before
						count = 3
						if realm != "" {
							count++
						}
					case "period.commit":
						runner.failure = "period.commit"
						count = 6
					case "published_policy_check":
						runner.bodies["published_policy_check"] = "{}"
						count = 9
					}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_pipe_update", Parameters: params})
					if scenario == "success" || scenario == "priority" {
						if err != nil {
							t.Fatal(err)
						}
					} else {
						var ae *cephdomain.ActionError
						if !errors.As(err, &ae) || ae.Retryable {
							t.Fatalf("error %v", err)
						}
					}
					if len(runner.calls) != count {
						t.Fatalf("calls %+v want %d", runner.calls, count)
					}
					for _, call := range runner.calls {
						if call.ID == "rgw_zonegroup.sync_pipe_update.write" {
							args := []string{"sync", "group", "pipe", "modify", "--group-id", "g", "--pipe-id", "p", "--source-tenant", "team", "--source-bucket", "new", "--source-bucket-id", "id", "--dest-tenant", "", "--dest-bucket", "archive", "--dest-bucket-id", "*", "--mode", mode}
							if mode == "user" {
								args = append(args, "--uid", "new$u")
							}
							if scenario == "priority" {
								args = append(args, "--priority", "-2147483648")
							}
							args = append(args, "--zonegroup-id", "zg", "--format", "json")
							if !call.Mutating || !reflect.DeepEqual(call.Args, args) {
								t.Fatalf("write %+v want %+v", call, args)
							}
						}
						if count <= 1 && call.Mutating {
							t.Fatal("preflight wrote")
						}
					}
				})
			}
		}
	}
}
