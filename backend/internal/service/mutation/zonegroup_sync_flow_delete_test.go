package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestZonegroupSyncFlowDeletion(t *testing.T) {
	for _, realm := range []string{"", "realm"} {
		for _, kind := range []string{"symmetrical", "directional"} {
			for _, keep := range []bool{false, true} {
				for _, scenario := range []string{"success", "missing", "duplicate", "stale", "read failure", "write failure", "readback mismatch", "period failure", "published mismatch"} {
					if realm == "" && (scenario == "period failure" || scenario == "published mismatch") {
						continue
					}
					t.Run(realm+"/"+kind+"/"+scenario+map[bool]string{true: "/keep", false: "/last"}[keep], func(t *testing.T) {
						service, _, cluster := newCephUserService(t)
						target := map[string]any{"id": "f", "zones": []any{"orphan", "z"}}
						other := map[string]any{"id": "other", "zones": []any{"z"}}
						params := map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "flow_type": kind}
						args := []string{"sync", "group", "flow", "remove", "--group-id", "g", "--flow-type", kind}
						if kind == "symmetrical" {
							params["flow_id"] = "f"
							args = append(args, "--flow-id", "f")
						} else {
							target = map[string]any{"source_zone": "orphan", "dest_zone": "z"}
							other = map[string]any{"source_zone": "z", "dest_zone": "orphan"}
							params["source_zone"] = "orphan"
							params["dest_zone"] = "z"
							args = append(args, "--flow-id", "directional", "--source-zone-id", "orphan", "--dest-zone-id", "z")
						}
						args = append(args, "--zonegroup-id", "zg", "--format", "json")
						entries := []any{target}
						if keep {
							entries = append(entries, other)
						}
						if scenario == "missing" {
							entries = []any{other}
						}
						if scenario == "duplicate" {
							entries = append(entries, target)
						}
						data := map[string]any{kind: entries}
						group := map[string]any{"id": "g", "status": "forbidden", "data_flow": data, "pipes": []any{map[string]any{"id": "p", "priority": json.Number("9007199254740993")}}}
						encode := func(v any) string {
							b, e := json.Marshal(v)
							if e != nil {
								t.Fatal(e)
							}
							return string(b)
						}
						params["expected_group"] = encode(group)
						// Orphan IDs remain deletable without an ID-to-name lookup.
						zg := map[string]any{"id": "zg", "name": "east", "realm_id": realm, "zones": []any{}, "sync_policy": map[string]any{"groups": []any{group}}}
						before := encode(zg)
						if keep {
							data[kind] = []any{other}
						} else {
							delete(data, kind)
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
						case "read failure":
							runner.failure = "pre_check"
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
						_, err := service.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.sync_flow_delete", Parameters: params})
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
							if call.ID == "rgw_zonegroup.sync_flow_delete.write" && (!call.Mutating || !reflect.DeepEqual(call.Args, args)) {
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
}
