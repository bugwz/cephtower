package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestZonegroupSyncGroupCreation(t *testing.T) {
	old := `{"id":"z-old","status":"forbidden","data_flow":{},"pipes":[{"id":"p","priority":9007199254740993}]}`
	for _, realm := range []string{"", "realm"} {
		for _, status := range []string{"enabled", "allowed", "forbidden"} {
			for _, existing := range []string{"", old} {
				policy := `{"groups":[` + existing + `]}`
				newGroup := `{"id":"a-new","status":"` + status + `","data_flow":{},"pipes":[]}`
				extra := ""
				if existing != "" {
					extra = "," + existing
				}
				changed := `{"groups":[` + newGroup + extra + `]}`
				before := `{"id":"zg","name":"east","realm_id":"` + realm + `","sync_policy":` + policy + `}`
				after := strings.Replace(before, policy, changed, 1)
				period := `{"id":"p","realm_id":"realm","epoch":2,"period_map":{"zonegroups":[{"id":"zg","sync_policy":` + changed + `}]}}`
				for _, tc := range []struct {
					name, failure, before, expected, after, published string
					count                                             int
				}{
					{name: "create"},
					{name: "exists", before: after, expected: changed, count: 1},
					{name: "stale", expected: changed, count: 1},
					{name: "bad expectation", expected: "broken", count: 1},
					{name: "write failure", failure: "write", count: 2},
					{name: "post mismatch", after: before, count: 3},
					{name: "period failure", failure: "period.commit", count: 5},
					{name: "published mismatch", published: strings.Replace(period, `"a-new"`, `"other"`, 1), count: 8},
				} {
					if realm == "" && (tc.name == "period failure" || tc.name == "published mismatch") {
						continue
					}
					t.Run(realm+"/"+status+"/"+existing+"/"+tc.name, func(t *testing.T) {
						initial := tc.before
						if initial == "" {
							initial = before
						}
						expected := tc.expected
						if expected == "" {
							expected = policy
						}
						actual := tc.after
						if actual == "" {
							actual = after
						}
						service, _, id := newCephUserService(t)
						runner := &zonegroupSyncExecutor{failure: tc.failure, bodies: map[string]string{"pre_check": initial, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "post_check": actual, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
						service.executor = runner
						if tc.published != "" {
							runner.bodies["published_policy_check"] = tc.published
						}
						_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_zonegroup.sync_group_create", Parameters: map[string]any{"name": "east", "zonegroup_id": "zg", "realm_id": realm, "group_id": "a-new", "status": status, "expected_policy": expected}})
						count := tc.count
						if count == 0 {
							count = 3
							if realm != "" {
								count = 9
							}
						} else if count > 1 && realm != "" {
							count++
						}
						if tc.name == "create" {
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
							t.Fatalf("calls %v", runner.calls)
						}
						for _, call := range runner.calls {
							if call.ID == "rgw_zonegroup.sync_group_create.write" {
								want := []string{"sync", "group", "create", "--group-id", "a-new", "--status", status, "--zonegroup-id", "zg", "--format", "json"}
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
}
