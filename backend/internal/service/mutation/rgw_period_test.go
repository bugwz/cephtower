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

type periodExecutor struct {
	calls   []executor.CommandSpec
	bodies  map[string]string
	failure string
	codes   map[string]int
	results []executor.CommandResult
}

func (e *periodExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_period.commit.")
	result := executor.CommandResult{Stdout: []byte(e.bodies[stage]), Stderr: []byte("private diagnostic"), ExitCode: e.codes[stage]}
	e.results = append(e.results, result)
	if stage == e.failure {
		return result, errors.New("private diagnostic")
	}
	return result, nil
}

func TestPeriodCommitRejectsExitCodesAndClearsOutputs(t *testing.T) {
	period := `{"id":"new","realm_id":"realm","epoch":2}`
	stages := []string{"pre_check", "commit", "realm_post_check", "period_post_check"}
	for index, stage := range stages {
		for _, mode := range []string{"exit", "error", "invalid_json"} {
			t.Run(stage+"/"+mode, func(t *testing.T) {
				service, _, id := newCephUserService(t)
				runner := &periodExecutor{codes: map[string]int{}, bodies: map[string]string{
					"pre_check": `{"id":"realm","current_period":"old"}`,
					"commit":    period, "period_post_check": period,
					"realm_post_check": `{"id":"realm","current_period":"new"}`,
				}}
				switch mode {
				case "exit":
					runner.codes[stage] = 5
				case "error":
					runner.failure = stage
				case "invalid_json":
					runner.bodies[stage] = `{"private":"diagnostic"} trailing`
				}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_period.commit", Parameters: map[string]any{"realm_id": "realm", "expected_current_period": "old"}})
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe error: %v", err)
				}
				if len(runner.calls) != index+1 {
					t.Fatalf("continued past failure: %v", runner.calls)
				}
				for _, result := range runner.results {
					for _, output := range [][]byte{result.Stdout, result.Stderr} {
						for _, b := range output {
							if b != 0 {
								t.Fatal("command output retained")
							}
						}
					}
				}
			})
		}
	}
}
func TestScopedPeriodCommit(t *testing.T) {
	period := `{"id":"new","realm_id":"realm","epoch":2,"period_map":{"zonegroups":[]}}`
	for _, tc := range []struct {
		name, stage, body, fail, code string
		count                         int
	}{
		{"success", "", "", "", "", 4},
		{"same period new epoch", "commit", strings.Replace(period, "new", "old", 1), "", "", 4},
		{"stale", "pre_check", `{"id":"realm","current_period":"changed"}`, "", "pre_check_failed", 1},
		{"wrong realm", "pre_check", `{"id":"other","current_period":"old"}`, "", "pre_check_failed", 1},
		{"read error", "", "", "pre_check", "pre_check_failed", 1},
		{"write error", "", "", "commit", "command_failed", 2},
		{"invalid commit", "commit", `broken`, "", "post_check_failed", 2},
		{"wrong result realm", "commit", strings.Replace(period, `"realm_id":"realm"`, `"realm_id":"other"`, 1), "", "post_check_failed", 2},
		{"invalid epoch", "commit", strings.Replace(period, `"epoch":2`, `"epoch":0`, 1), "", "post_check_failed", 2},
		{"realm read error", "", "", "realm_post_check", "post_check_failed", 3},
		{"current mismatch", "realm_post_check", `{"id":"realm","current_period":"other"}`, "", "post_check_failed", 3},
		{"period read error", "", "", "period_post_check", "post_check_failed", 4},
		{"period drift", "period_post_check", strings.Replace(period, `"epoch":2`, `"epoch":3`, 1), "", "post_check_failed", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &periodExecutor{failure: tc.fail, bodies: map[string]string{"pre_check": `{"id":"realm","current_period":"old"}`, "commit": period, "realm_post_check": `{"id":"realm","current_period":"new"}`, "period_post_check": period}}
			if tc.stage != "" {
				runner.bodies[tc.stage] = tc.body
			}
			if tc.name == "same period new epoch" {
				runner.bodies["realm_post_check"] = `{"id":"realm","current_period":"old"}`
				runner.bodies["period_post_check"] = tc.body
			}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_period.commit", Parameters: map[string]any{"realm_id": "realm", "expected_current_period": "old"}})
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Code != tc.code || ae.Retryable {
					t.Fatalf("error %v", err)
				}
			}
			if len(runner.calls) != tc.count {
				t.Fatalf("calls: %v", runner.calls)
			}
			for i, call := range runner.calls {
				want := []string{"realm", "get", "--realm-id", "realm", "--format", "json"}
				if i == 1 {
					want = []string{"period", "update", "--commit", "--realm-id", "realm", "--format", "json"}
				}
				if i == 3 {
					want = []string{"period", "get", "--realm-id", "realm", "--format", "json"}
				}
				if call.Mutating != (i == 1) || !reflect.DeepEqual(call.Args, want) {
					t.Fatalf("command %+v", call)
				}
			}
		})
	}
	for _, p := range []map[string]any{nil, {"realm_id": "realm"}, {"realm_id": "-bad", "expected_current_period": "old"}, {"realm_id": "realm", "expected_current_period": ""}} {
		if _, err := build(Request{Action: "rgw_period.commit"}, p); err == nil {
			t.Fatalf("accepted %v", p)
		}
	}
}
