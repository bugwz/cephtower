package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupSyncPipeCreation(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, mode := range []string{"system", "user"} {
			for _, scenario := range []string{"success", "wildcard", "existing", "stale", "unknown zone", "mixed wildcard", "write", "post_check", "period.commit", "published_policy_check"} {
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
					old := map[string]any{"id": "old", "params": map[string]any{"priority": json.Number("9007199254740993")}}
					group := map[string]any{"id": "g", "status": "allowed", "data_flow": map[string]any{}, "pipes": []any{old}}
					params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "pipe_id": "p", "source_zones": []any{"z2", "z1"}, "dest_zones": []any{"*"}, "source_tenant": "team", "source_bucket": "photos", "source_bucket_id": "instance", "dest_tenant": "*", "dest_bucket": "*", "mode": mode}
					if mode == "user" {
						params["user"] = "team$u"
					}
					sourceZones := []any{"z1", "z2"}
					sourceArg := "z1,z2"
					if scenario == "wildcard" {
						params["source_zones"] = []any{"*"}
						sourceZones = []any{"*"}
						sourceArg = "*"
					}
					if scenario == "existing" {
						params["pipe_id"] = "old"
					}
					expected := encode(group)
					params["expected_group"] = expected
					zg := map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{map[string]any{"id": "z1", "name": "west"}, map[string]any{"id": "z2", "name": "east"}}, "sync_policy": map[string]any{"groups": []any{group}}}
					before := encode(zg)
					pipeParams := map[string]any{"source": map[string]any{"filter": map[string]any{"tags": []any{}}}, "dest": map[string]any{}, "priority": json.Number("0"), "mode": mode}
					if mode == "user" {
						pipeParams["user"] = "team$u"
					}
					added := map[string]any{"id": "p", "source": map[string]any{"bucket": "team/photos:instance", "zones": sourceZones}, "dest": map[string]any{"bucket": "*", "zones": []any{"*"}}, "params": pipeParams}
					group["pipes"] = []any{old, added}
					after := encode(zg)
					period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{zg}}})
					runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": before, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "post_check": after, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
					count := 3
					if realm != "" {
						count = 9
					}
					switch scenario {
					case "existing":
						count = 1
					case "stale":
						params["expected_group"] = encode(group)
						count = 1
					case "unknown zone":
						params["source_zones"] = []any{"missing"}
						count = 1
					case "mixed wildcard":
						params["source_zones"] = []any{"*", "z1"}
						count = 0
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
					_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_pipe_create", Parameters: params})
					if scenario == "success" || scenario == "wildcard" {
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
						if call.ID == "rgw_zonegroup.sync_pipe_create.write" {
							args := []string{"sync", "group", "pipe", "create", "--group-id", "g", "--pipe-id", "p", "--source-zone-ids", sourceArg, "--source-tenant", "team", "--source-bucket", "photos", "--source-bucket-id", "instance", "--dest-zone-ids", "*", "--dest-tenant", "*", "--dest-bucket", "*", "--dest-bucket-id", "*", "--mode", mode}
							if mode == "user" {
								args = append(args, "--uid", "team$u")
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
