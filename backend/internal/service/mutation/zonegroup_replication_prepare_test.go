package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupReplicationPreparation(t *testing.T) {
	for _, others := range []bool{false, true} {
		for _, scenario := range []string{"success", "existing", "stale", "changed zones", "no realm", "no zones", "duplicate zones", "realm_pre_check", "group", "group_post_check", "flow", "flow_post_check", "pipe", "pipe_post_check", "period.commit", "published_policy_check"} {
			t.Run(scenario+map[bool]string{false: "/empty", true: "/others"}[others], func(t *testing.T) {
				service, _, cluster := newCephUserService(t)
				encode := func(v any) string {
					b, e := json.Marshal(v)
					if e != nil {
						t.Fatal(e)
					}
					return string(b)
				}
				old := []any{}
				if others {
					old = []any{map[string]any{"id": "z", "status": "forbidden", "data_flow": map[string]any{}, "pipes": []any{map[string]any{"id": "old", "priority": json.Number("9007199254740993")}}}}
				}
				policy := map[string]any{"groups": old}
				zg := map[string]any{"id": "zg", "name": "east", "realm_id": "realm", "zones": []any{map[string]any{"id": "z2", "name": "east"}, map[string]any{"id": "z1", "name": "west"}}, "sync_policy": policy}
				params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": "realm", "expected_policy": encode(policy), "expected_zones": []any{"z2", "z1"}}
				if scenario == "no realm" {
					params["realm_id"] = ""
					zg["realm_id"] = ""
				}
				if scenario == "no zones" {
					zg["zones"] = []any{}
				}
				if scenario == "duplicate zones" {
					zg["zones"] = []any{map[string]any{"id": "z1", "name": "a"}, map[string]any{"id": "z1", "name": "b"}}
				}
				group := map[string]any{"id": "dashboard_admin_group", "status": "allowed", "data_flow": map[string]any{}, "pipes": []any{}}
				if scenario == "existing" {
					policy["groups"] = append([]any{group}, old...)
					params["expected_policy"] = encode(policy)
				}
				before := encode(zg)
				if scenario == "stale" {
					params["expected_policy"] = `{"groups":[{"id":"different","status":"allowed","data_flow":{},"pipes":[]}]}`
				}
				if scenario == "changed zones" {
					params["expected_zones"] = []any{"z1"}
				}
				policy["groups"] = append([]any{group}, old...)
				groupResult := encode(zg)
				group["data_flow"] = map[string]any{"symmetrical": []any{map[string]any{"id": "dashboard_admin_flow", "zones": []any{"z1", "z2"}}}}
				flowResult := encode(zg)
				group["pipes"] = []any{map[string]any{"id": "dashboard_admin_pipe", "source": map[string]any{"bucket": "*", "zones": []any{"*"}}, "dest": map[string]any{"bucket": "*", "zones": []any{"*"}}, "params": map[string]any{"mode": "system", "priority": json.Number("0"), "source": map[string]any{"filter": map[string]any{"tags": []any{}}}, "dest": map[string]any{}}}}
				pipeResult := encode(zg)
				period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{zg}}})
				runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": before, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "group_post_check": groupResult, "flow_post_check": flowResult, "pipe_post_check": pipeResult, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
				stages := []string{"pre_check", "realm_pre_check", "group", "group_post_check", "flow", "flow_post_check", "pipe", "pipe_post_check", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check", "published_policy_check"}
				count := len(stages)
				switch scenario {
				case "success":
				case "existing", "stale", "changed zones", "no realm", "no zones", "duplicate zones":
					count = 1
				default:
					for i, stage := range stages {
						if stage == scenario {
							count = i + 1
						}
					}
					if scenario == "group_post_check" || scenario == "flow_post_check" || scenario == "pipe_post_check" || scenario == "published_policy_check" {
						runner.bodies[scenario] = "{}"
					} else {
						runner.failure = scenario
					}
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.replication_prepare", Parameters: params})
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
				commands := map[string][]string{
					"group": {"sync", "group", "create", "--group-id", "dashboard_admin_group", "--status", "allowed", "--zonegroup-id", "zg", "--format", "json"},
					"flow":  {"sync", "group", "flow", "create", "--group-id", "dashboard_admin_group", "--flow-type", "symmetrical", "--flow-id", "dashboard_admin_flow", "--zone-ids", "z1,z2", "--zonegroup-id", "zg", "--format", "json"},
					"pipe":  {"sync", "group", "pipe", "create", "--group-id", "dashboard_admin_group", "--pipe-id", "dashboard_admin_pipe", "--source-zone-ids", "*", "--source-tenant", "", "--source-bucket", "*", "--source-bucket-id", "*", "--dest-zone-ids", "*", "--dest-tenant", "", "--dest-bucket", "*", "--dest-bucket-id", "*", "--mode", "system", "--zonegroup-id", "zg", "--format", "json"},
				}
				for i, call := range runner.calls {
					if call.ID != "rgw_zonegroup.replication_prepare."+stages[i] {
						t.Fatalf("stage %+v", call)
					}
					if want, ok := commands[stages[i]]; ok && (!call.Mutating || !reflect.DeepEqual(call.Args, want)) {
						t.Fatalf("command %+v want %+v", call, want)
					}
					if count <= 2 && call.Mutating {
						t.Fatal("preflight wrote")
					}
				}
			})
		}
	}
}
