package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupSyncFlowMembership(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, mode := range []string{"add", "remove", "replace"} {
			for _, scenario := range []string{"success", "missing", "duplicate", "stale", "unchanged", "unknown zone", "empty", "add", "add_post_check", "remove", "remove_post_check", "period.commit", "published_policy_check"} {
				if mode == "add" && (scenario == "remove" || scenario == "remove_post_check") {
					continue
				}
				if mode == "remove" && (scenario == "add" || scenario == "add_post_check") {
					continue
				}
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
					makeGroup := func(zones []any) map[string]any {
						return map[string]any{"id": "g", "status": "allowed", "data_flow": map[string]any{"symmetrical": []any{map[string]any{"id": "f", "zones": zones}, map[string]any{"id": "other", "zones": []any{"b"}}}, "directional": []any{map[string]any{"source_zone": "a", "dest_zone": "b"}}}, "pipes": []any{map[string]any{"priority": json.Number("9007199254740993")}}}
					}
					makeZonegroup := func(group map[string]any) map[string]any {
						return map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{map[string]any{"id": "a", "name": "west"}, map[string]any{"id": "b", "name": "east"}}, "sync_policy": map[string]any{"groups": []any{group}}}
					}
					old := []any{"a"}
					desired := []any{"a", "b"}
					intermediate := []any{"a", "b"}
					if mode == "remove" {
						old = []any{"a", "b"}
						desired = []any{"b"}
					}
					if mode == "replace" {
						old = []any{"orphan"}
						desired = []any{"b"}
						intermediate = []any{"b", "orphan"}
					}
					group := makeGroup(old)
					params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "flow_id": "f", "zones": desired, "expected_group": encode(group)}
					if scenario == "missing" {
						params["flow_id"] = "missing"
					}
					if scenario == "duplicate" {
						data := group["data_flow"].(map[string]any)
						data["symmetrical"] = append(data["symmetrical"].([]any), map[string]any{"id": "f", "zones": old})
						params["expected_group"] = encode(group)
					}
					if scenario == "stale" {
						params["expected_group"] = encode(makeGroup([]any{"bad"}))
					}
					if scenario == "unchanged" {
						params["zones"] = old
					}
					if scenario == "unknown zone" {
						params["zones"] = []any{"missing"}
					}
					if scenario == "empty" {
						params["zones"] = []any{}
					}
					final := makeZonegroup(makeGroup(desired))
					period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{final}}})
					runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": encode(makeZonegroup(group)), "realm_pre_check": `{"id":"realm","current_period":"p"}`, "add_post_check": encode(makeZonegroup(makeGroup(intermediate))), "remove_post_check": encode(final), "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
					stages := []string{"pre_check"}
					if realm != "" {
						stages = append(stages, "realm_pre_check")
					}
					if mode != "remove" {
						stages = append(stages, "add", "add_post_check")
					}
					if mode != "add" {
						stages = append(stages, "remove", "remove_post_check")
					}
					if realm != "" {
						stages = append(stages, "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check", "published_policy_check")
					}
					count := len(stages)
					switch scenario {
					case "missing", "duplicate", "stale", "unchanged", "unknown zone":
						count = 1
					case "empty":
						count = 0
					case "success":
					default:
						for i, stage := range stages {
							if stage == scenario {
								count = i + 1
							}
						}
						if scenario == "add_post_check" || scenario == "remove_post_check" || scenario == "published_policy_check" {
							runner.bodies[scenario] = "{}"
						} else {
							runner.failure = scenario
						}
					}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_flow_update", Parameters: params})
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
					for i, call := range runner.calls {
						if call.ID != "rgw_zonegroup.sync_flow_update."+stages[i] {
							t.Fatalf("stage %+v", call)
						}
						if call.Mutating && (stages[i] == "add" || stages[i] == "remove") {
							verb, ids := "create", "b"
							if stages[i] == "remove" {
								verb = "remove"
								ids = "a"
								if mode == "replace" {
									ids = "orphan"
								}
							}
							want := []string{"sync", "group", "flow", verb, "--group-id", "g", "--flow-type", "symmetrical", "--flow-id", "f", "--zone-ids", ids, "--zonegroup-id", "zg", "--format", "json"}
							if !reflect.DeepEqual(call.Args, want) {
								t.Fatalf("args %+v want %+v", call.Args, want)
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
