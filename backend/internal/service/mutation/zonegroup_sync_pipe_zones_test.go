package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupPipeZoneMembership(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, tc := range []struct {
			name                                                       string
			oldSource, oldDest, source, dest, middleSource, middleDest []any
			add, remove                                                []string
		}{
			{"replace", []any{"a"}, []any{"*"}, []any{"b"}, []any{"c"}, []any{"a", "b"}, []any{"c"}, []string{"--source-zone-ids", "b", "--dest-zone-ids", "c"}, []string{"--source-zone-ids", "a"}},
			{"to wildcard", []any{"a"}, []any{"b"}, []any{"*"}, []any{"*"}, []any{"*"}, []any{"*"}, []string{"--source-zone-ids", "*", "--dest-zone-ids", "*"}, nil},
			{"from wildcard", []any{"*"}, []any{"*"}, []any{"a", "b"}, []any{"b"}, []any{"a", "b"}, []any{"b"}, []string{"--source-zone-ids", "a,b", "--dest-zone-ids", "b"}, nil},
			{"remove only", []any{"a", "b"}, []any{"b", "c"}, []any{"b"}, []any{"b"}, nil, nil, nil, []string{"--source-zone-ids", "a", "--dest-zone-ids", "c"}},
			{"empty to explicit", []any{}, []any{"a"}, []any{"b"}, []any{"a", "b"}, []any{"b"}, []any{"a", "b"}, []string{"--source-zone-ids", "b", "--dest-zone-ids", "b"}, nil},
			{"orphan", []any{"orphan"}, []any{"*"}, []any{"a"}, []any{"*"}, []any{"a", "orphan"}, []any{"*"}, []string{"--source-zone-ids", "a"}, []string{"--source-zone-ids", "orphan"}},
		} {
			for _, scenario := range []string{"success", "stale", "missing", "unknown zone", "empty", "add", "add_post_check", "remove", "remove_post_check", "period.commit", "published_policy_check"} {
				if (scenario == "add" || scenario == "add_post_check") && len(tc.add) == 0 {
					continue
				}
				if (scenario == "remove" || scenario == "remove_post_check") && len(tc.remove) == 0 {
					continue
				}
				if realm == "" && (scenario == "period.commit" || scenario == "published_policy_check") {
					continue
				}
				t.Run(realm+"/"+tc.name+"/"+scenario, func(t *testing.T) {
					service, _, cluster := newCephUserService(t)
					encode := func(v any) string {
						b, e := json.Marshal(v)
						if e != nil {
							t.Fatal(e)
						}
						return string(b)
					}
					makeGroup := func(source, dest []any) map[string]any {
						return map[string]any{"id": "g", "status": "allowed", "data_flow": map[string]any{}, "pipes": []any{map[string]any{"id": "p", "source": map[string]any{"bucket": "team/photos:instance", "zones": source}, "dest": map[string]any{"bucket": "*", "zones": dest}, "params": map[string]any{"priority": json.Number("9007199254740993"), "mode": "user", "user": "team$u", "filter": map[string]any{"prefix": "x"}}}}}
					}
					makeZG := func(group map[string]any) map[string]any {
						return map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{map[string]any{"id": "a", "name": "west"}, map[string]any{"id": "b", "name": "east"}, map[string]any{"id": "c", "name": "north"}}, "sync_policy": map[string]any{"groups": []any{group}}}
					}
					group := makeGroup(tc.oldSource, tc.oldDest)
					params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "pipe_id": "p", "source_zones": tc.source, "dest_zones": tc.dest, "expected_group": encode(group)}
					final := makeZG(makeGroup(tc.source, tc.dest))
					period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{final}}})
					runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": encode(makeZG(group)), "realm_pre_check": `{"id":"realm","current_period":"p"}`, "add_post_check": encode(makeZG(makeGroup(tc.middleSource, tc.middleDest))), "remove_post_check": encode(final), "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
					stages := []string{"pre_check"}
					if realm != "" {
						stages = append(stages, "realm_pre_check")
					}
					if len(tc.add) > 0 {
						stages = append(stages, "add", "add_post_check")
					}
					if len(tc.remove) > 0 {
						stages = append(stages, "remove", "remove_post_check")
					}
					if realm != "" {
						stages = append(stages, "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check", "published_policy_check")
					}
					count := len(stages)
					switch scenario {
					case "success":
					case "stale":
						params["expected_group"] = encode(makeGroup([]any{"different"}, tc.oldDest))
						count = 1
					case "missing":
						params["pipe_id"] = "missing"
						count = 1
					case "unknown zone":
						params["source_zones"] = []any{"unknown"}
						count = 1
					case "empty":
						params["source_zones"] = []any{}
						count = 0
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
					_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_pipe_zones", Parameters: params})
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
						stage := stages[i]
						if call.ID != "rgw_zonegroup.sync_pipe_zones."+stage {
							t.Fatalf("stage %+v", call)
						}
						if stage == "add" || stage == "remove" {
							verb, flags := "modify", tc.add
							if stage == "remove" {
								verb, flags = "remove", tc.remove
							}
							args := append([]string{"sync", "group", "pipe", verb, "--group-id", "g", "--pipe-id", "p", "--zonegroup-id", "zg", "--format", "json"}, flags...)
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

func TestNativePipeZonePlannerValidation(t *testing.T) {
	for _, tc := range []struct {
		old, desired []any
		body         string
		valid        bool
	}{
		{[]any{"a"}, []any{"b"}, `{"zones":[{"id":"a"},{"id":"b"}]}`, true},
		{[]any{"orphan"}, []any{"*"}, `{"zones":[]}`, true},
		{[]any{"a", "a"}, []any{"b"}, `{"zones":[{"id":"b"}]}`, false},
		{[]any{"*", "a"}, []any{"b"}, `{"zones":[{"id":"b"}]}`, false},
		{[]any{"a"}, []any{"*", "b"}, `{"zones":[{"id":"b"}]}`, false},
		{[]any{"a"}, []any{"b"}, `{"zones":[{"id":"b"},{"id":"b"}]}`, false},
		{[]any{"a"}, []any{"*"}, `{}`, false},
	} {
		_, err := planSyncPipeZones(map[string]any{"zones": tc.old}, tc.desired, []byte(tc.body), true)
		if (err == nil) != tc.valid {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
}
