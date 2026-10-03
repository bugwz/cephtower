package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type zonegroupSyncExecutor struct {
	calls   []executor.CommandSpec
	bodies  map[string]string
	failure string
}

func (e *zonegroupSyncExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(strings.TrimPrefix(spec.ID, "rgw_zonegroup.sync_group."), "rgw_zonegroup.sync_group_create.")
	stage = strings.TrimPrefix(stage, "rgw_zonegroup.sync_group_delete.")
	if stage == e.failure {
		return executor.CommandResult{}, errors.New("injected failure")
	}
	return executor.CommandResult{Stdout: []byte(e.bodies[stage])}, nil
}
func TestZonegroupSyncStatusAndPublication(t *testing.T) {
	group := `{"id":"g","status":"allowed","data_flow":{},"pipes":[{"id":"p","priority":9007199254740993}]}`
	for _, realm := range []string{"", "realm"} {
		for _, status := range []string{"enabled", "forbidden"} {
			policy := `{"groups":[` + group + `]}`
			changedPolicy := strings.Replace(policy, `"allowed"`, `"`+status+`"`, 1)
			before := `{"id":"zg","name":"east","realm_id":"` + realm + `","sync_policy":` + policy + `,"endpoints":["keep"]}`
			after := strings.Replace(before, policy, changedPolicy, 1)
			period := `{"id":"period","realm_id":"realm","epoch":2,"period_map":{"zonegroups":[{"id":"zg","sync_policy":` + changedPolicy + `}]}}`
			stages := []string{"pre_check", "write", "post_check"}
			if realm != "" {
				stages = []string{"pre_check", "realm_pre_check", "write", "post_check", "period.pre_check", "period.commit", "period.realm_post_check", "period.period_post_check", "published_policy_check"}
			}
			failures := append([]string{""}, stages...)
			for _, failure := range failures {
				t.Run(realm+"/"+status+"/"+failure, func(t *testing.T) {
					service, _, id := newCephUserService(t)
					runner := &zonegroupSyncExecutor{failure: failure, bodies: map[string]string{"pre_check": before, "realm_pre_check": `{"id":"realm","current_period":"period"}`, "post_check": after, "period.pre_check": `{"id":"realm","current_period":"period"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"period"}`, "period.period_post_check": period, "published_policy_check": period}}
					service.executor = runner
					_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_zonegroup.sync_group", Parameters: map[string]any{"zonegroup_id": "zg", "name": "east", "realm_id": realm, "group_id": "g", "expected_group": group, "status": status}})
					if failure == "" {
						if err != nil {
							t.Fatal(err)
						}
					} else {
						var ae *cephdomain.ActionError
						if !errors.As(err, &ae) || ae.Retryable {
							t.Fatalf("error %v", err)
						}
					}
					wantCount := len(stages)
					for i, stage := range stages {
						if stage == failure {
							wantCount = i + 1
							break
						}
					}
					if len(runner.calls) != wantCount {
						t.Fatalf("calls %v", runner.calls)
					}
					for i, call := range runner.calls {
						stage := stages[i]
						if call.ID != "rgw_zonegroup.sync_group."+stage || call.Mutating != (stage == "write" || stage == "period.commit") {
							t.Fatalf("call %+v", call)
						}
						if stage == "write" {
							want := []string{"sync", "group", "modify", "--group-id", "g", "--status", status, "--zonegroup-id", "zg", "--format", "json"}
							if !reflect.DeepEqual(call.Args, want) {
								t.Fatalf("args %v", call.Args)
							}
						}
						if strings.HasPrefix(stage, "period.") || stage == "realm_pre_check" || stage == "published_policy_check" {
							hasRealm := false
							for j, arg := range call.Args {
								if arg == "--realm-id" && j+1 < len(call.Args) && call.Args[j+1] == "realm" {
									hasRealm = true
								}
							}
							if !hasRealm {
								t.Fatalf("unscoped period call %+v", call)
							}
						}
					}
				})
			}
		}
	}
}

func TestZonegroupSyncMismatchesStop(t *testing.T) {
	group := `{"id":"g","status":"allowed","data_flow":{},"pipes":[]}`
	before := `{"id":"zg","name":"east","realm_id":"realm","sync_policy":{"groups":[` + group + `]}}`
	after := strings.Replace(before, `"allowed"`, `"enabled"`, 1)
	period := `{"id":"p","realm_id":"realm","epoch":2,"period_map":{"zonegroups":[{"id":"zg","sync_policy":{"groups":[` + strings.Replace(group, `"allowed"`, `"enabled"`, 1) + `]}}]}}`
	for _, tc := range []struct {
		name, stage, body string
		count             int
	}{
		{"wrong zonegroup", "pre_check", strings.Replace(before, `"zg"`, `"other"`, 1), 1},
		{"wrong realm", "pre_check", strings.Replace(before, `"realm_id":"realm"`, `"realm_id":"other"`, 1), 1},
		{"group stale", "pre_check", after, 1},
		{"post unchanged", "post_check", before, 4},
		{"unrelated change", "post_check", strings.Replace(after, `"east"`, `"west"`, 1), 4},
		{"realm changed after write", "period.pre_check", `{"id":"realm","current_period":"new"}`, 5},
		{"published wrong policy", "published_policy_check", strings.Replace(period, `"enabled"`, `"allowed"`, 1), 9},
		{"published missing group", "published_policy_check", `{"realm_id":"realm","period_map":{"zonegroups":[]}}`, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &zonegroupSyncExecutor{bodies: map[string]string{"pre_check": before, "realm_pre_check": `{"id":"realm","current_period":"p"}`, "post_check": after, "period.pre_check": `{"id":"realm","current_period":"p"}`, "period.commit": period, "period.realm_post_check": `{"id":"realm","current_period":"p"}`, "period.period_post_check": period, "published_policy_check": period}}
			runner.bodies[tc.stage] = tc.body
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_zonegroup.sync_group", Parameters: map[string]any{"zonegroup_id": "zg", "name": "east", "realm_id": "realm", "group_id": "g", "expected_group": group, "status": "enabled"}})
			var ae *cephdomain.ActionError
			if !errors.As(err, &ae) || ae.Retryable || len(runner.calls) != tc.count {
				t.Fatalf("error %v calls %v", err, runner.calls)
			}
		})
	}
}
