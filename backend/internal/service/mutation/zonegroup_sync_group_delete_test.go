package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestZonegroupSyncGroupDeletion(t *testing.T) {
	group := `{"id":"g","status":"forbidden","data_flow":{"symmetrical":[{"id":"f","zones":["z"]}]},"pipes":[{"id":"p","priority":9007199254740993}]}`
	other := `{"id":"other","status":"enabled","data_flow":{},"pipes":[]}`
	for _, realm := range []string{"", "realm"} {
		for _, remaining := range []string{"", other} {
			tail := ""
			if remaining != "" {
				tail = "," + remaining
			}
			before := `{"id":"zg","name":"east","realm_id":"` + realm + `","sync_policy":{"groups":[` + group + tail + `]}}`
			after := `{"id":"zg","name":"east","realm_id":"` + realm + `","sync_policy":{"groups":[` + remaining + `]}}`
			period := `{"id":"p","realm_id":"realm","epoch":2,"period_map":{"zonegroups":[{"id":"zg","sync_policy":{"groups":[` + remaining + `]}}]}}`
			for _, tc := range []struct {
				name, initial, expected, after, published, failure string
				count                                              int
			}{
				{name: "delete"},
				{name: "missing", initial: after, count: 1},
				{name: "stale", expected: strings.Replace(group, "forbidden", "allowed", 1), count: 1},
				{name: "malformed snapshot", expected: "broken", count: 1},
				{name: "duplicate group", initial: strings.Replace(before, group, group+","+group, 1), count: 1},
				{name: "read failure", failure: "pre_check", count: 1},
				{name: "write failure", failure: "write", count: 2},
				{name: "readback failure", failure: "post_check", count: 3},
				{name: "still exists", after: before, count: 3},
				{name: "unrelated change", after: strings.Replace(after, `"east"`, `"changed"`, 1), count: 3},
				{name: "period failure", failure: "period.commit", count: 5},
				{name: "published mismatch", published: strings.Replace(period, `"groups":[`+remaining+`]`, `"groups":[`+group+tail+`]`, 1), count: 8},
			} {
				if realm == "" && (tc.name == "period failure" || tc.name == "published mismatch") {
					continue
				}
				t.Run(realm+"/"+remaining+"/"+tc.name, func(t *testing.T) {
					service, _, id := newCephUserService(t)
					initial := tc.initial
					if initial == "" {
						initial = before
					}
					expected := tc.expected
					if expected == "" {
						expected = group
					}
					actual := tc.after
					if actual == "" {
						actual = after
					}
					published := tc.published
					if published == "" {
						published = period
					}
					runner := &zonegroupSyncExecutor{failure: tc.failure, bodies: map[string]string{"pre_check": initial, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "post_check": actual, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": published}}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_zonegroup.sync_group_delete", Parameters: map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "g", "expected_group": expected}})
					if tc.name == "delete" {
						if err != nil {
							t.Fatal(err)
						}
					} else {
						var ae *cephdomain.ActionError
						if !errors.As(err, &ae) || ae.Retryable {
							t.Fatalf("error %v", err)
						}
					}
					count := tc.count
					if count == 0 {
						count = 3
						if realm != "" {
							count = 9
						}
					} else if count > 1 && realm != "" {
						count++
					}
					if len(runner.calls) != count {
						t.Fatalf("calls: %+v", runner.calls)
					}
					for _, call := range runner.calls {
						if call.ID == "rgw_zonegroup.sync_group_delete.write" {
							want := []string{"sync", "group", "remove", "--group-id", "g", "--zonegroup-id", "zg", "--format", "json"}
							if !call.Mutating || !reflect.DeepEqual(call.Args, want) {
								t.Fatalf("write %+v", call)
							}
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
