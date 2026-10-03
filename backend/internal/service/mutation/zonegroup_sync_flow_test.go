package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestZonegroupSyncFlowCreation(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, kind := range []string{"symmetrical", "directional"} {
			for _, scenario := range []string{"success", "existing", "stale", "unknown zone", "duplicate zones", "write failure", "readback mismatch", "period failure", "published mismatch"} {
				if realm == "" && (scenario == "period failure" || scenario == "published mismatch") {
					continue
				}
				t.Run(realm+"/"+kind+"/"+scenario, func(t *testing.T) {
					service, _, cluster := newCephUserService(t)
					group := map[string]any{"id": "g", "status": "allowed", "data_flow": map[string]any{}, "pipes": []any{map[string]any{"priority": json.Number("9007199254740993")}}}
					params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "flow_type": kind}
					args := []string{"sync", "group", "flow", "create", "--group-id", "g", "--flow-type", kind}
					if kind == "symmetrical" {
						params["flow_id"] = "f"
						params["zones"] = []any{"z2", "z1"}
						args = append(args, "--flow-id", "f", "--zone-ids", "z1,z2")
					} else {
						params["source_zone"] = "z1"
						params["dest_zone"] = "z2"
						args = append(args, "--flow-id", "directional", "--source-zone-id", "z1", "--dest-zone-id", "z2")
					}
					args = append(args, "--zonegroup-id", "zg", "--format", "json")
					encode := func(v any) string {
						b, err := json.Marshal(v)
						if err != nil {
							t.Fatal(err)
						}
						return string(b)
					}
					zonegroup := map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{map[string]any{"id": "z1", "name": "west"}, map[string]any{"id": "z2", "name": "east"}}, "sync_policy": map[string]any{"groups": []any{group}}}
					if scenario == "existing" {
						if err := addBucketSyncFlow(group, params); err != nil {
							t.Fatal(err)
						}
					}
					params["expected_group"] = encode(group)
					before := encode(zonegroup)
					if scenario != "existing" {
						sorted := map[string]any{}
						for k, v := range params {
							sorted[k] = v
						}
						if kind == "symmetrical" {
							sorted["zones"] = []any{"z1", "z2"}
						}
						if err := addBucketSyncFlow(group, sorted); err != nil {
							t.Fatal(err)
						}
					}
					after := encode(zonegroup)
					if scenario != "existing" {
						wantFlow := `{"symmetrical":[{"id":"f","zones":["z1","z2"]}]}`
						if kind == "directional" {
							wantFlow = `{"directional":[{"source_zone":"z1","dest_zone":"z2"}]}`
						}
						if !reflect.DeepEqual(group["data_flow"], periodDocument([]byte(wantFlow))) {
							t.Fatalf("unexpected native flow: %+v", group["data_flow"])
						}
					}
					period := encode(map[string]any{"id": "p", "realm_id": "realm", "epoch": 2, "period_map": map[string]any{"zonegroups": []any{zonegroup}}})
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
						if kind == "symmetrical" {
							params["zones"] = []any{"missing"}
						} else {
							params["dest_zone"] = "missing"
						}
						count = 1
					case "duplicate zones":
						runner.bodies["pre_check"] = encode(map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{map[string]any{"id": "z1", "name": "x"}, map[string]any{"id": "z1", "name": "y"}}, "sync_policy": periodDocument([]byte(before))["sync_policy"]})
						count = 1
					case "write failure":
						runner.failure = "write"
						count = 2
						if realm != "" {
							count++
						}
					case "readback mismatch":
						runner.bodies["post_check"] = before
						count = 3
						if realm != "" {
							count++
						}
					case "period failure":
						runner.failure = "period.commit"
						count = 6
					case "published mismatch":
						runner.bodies["published_policy_check"] = `{"id":"p","realm_id":"realm","period_map":{"zonegroups":[]}}`
						count = 9
					}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_flow_create", Parameters: params})
					if (err == nil) != (scenario == "success") {
						t.Fatalf("error %v", err)
					}
					if len(runner.calls) != count {
						t.Fatalf("calls %d want %d: %+v", len(runner.calls), count, runner.calls)
					}
					for _, call := range runner.calls {
						if call.ID == "rgw_zonegroup.sync_flow_create.write" && (!call.Mutating || !reflect.DeepEqual(call.Args, args)) {
							t.Fatalf("write %+v", call)
						}
						if count == 1 && call.Mutating {
							t.Fatal("preflight wrote")
						}
					}
				})
			}
		}
	}
}
