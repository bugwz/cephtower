package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupSyncPipeDeletion(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, keep := range []bool{false, true} {
			for _, scenario := range []string{"success", "missing", "duplicate", "stale", "pre_check", "write", "post_check", "period.commit", "published_policy_check"} {
				if realm == "" && (scenario == "period.commit" || scenario == "published_policy_check") {
					continue
				}
				t.Run(realm+"/"+scenario+map[bool]string{false: "/last", true: "/keep"}[keep], func(t *testing.T) {
					service, _, cluster := newCephUserService(t)
					encode := func(v any) string {
						b, e := json.Marshal(v)
						if e != nil {
							t.Fatal(e)
						}
						return string(b)
					}
					target := map[string]any{"id": "p", "source": map[string]any{"bucket": "team/photos:instance", "zones": []any{"orphan"}}, "dest": map[string]any{"bucket": "*", "zones": []any{"*"}}, "params": map[string]any{"mode": "user", "user": "team$u", "priority": json.Number("9007199254740993"), "source": map[string]any{"filter": map[string]any{"prefix": "photos/"}}}}
					other := map[string]any{"id": "other", "params": map[string]any{"priority": json.Number("9007199254740993")}}
					pipes := []any{target}
					if keep {
						pipes = append(pipes, other)
					}
					if scenario == "missing" {
						pipes = []any{other}
					}
					if scenario == "duplicate" {
						pipes = append(pipes, target)
					}
					group := map[string]any{"id": "g", "status": "forbidden", "data_flow": map[string]any{"symmetrical": []any{map[string]any{"id": "flow", "zones": []any{"orphan"}}}}, "pipes": pipes}
					params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "pipe_id": "p", "expected_group": encode(group)}
					zg := map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{}, "sync_policy": map[string]any{"groups": []any{group}}}
					before := encode(zg)
					group["pipes"] = []any{}
					if keep {
						group["pipes"] = []any{other}
					}
					after := encode(zg)
					period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{zg}}})
					runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": before, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "post_check": after, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
					count := 3
					if realm != "" {
						count = 9
					}
					switch scenario {
					case "missing", "duplicate":
						count = 1
					case "stale":
						params["expected_group"] = encode(group)
						count = 1
					case "pre_check":
						runner.failure = "pre_check"
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
					_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_pipe_delete", Parameters: params})
					if scenario == "success" {
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
						if call.ID == "rgw_zonegroup.sync_pipe_delete.write" {
							args := []string{"sync", "group", "pipe", "remove", "--group-id", "g", "--pipe-id", "p", "--zonegroup-id", "zg", "--format", "json"}
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
