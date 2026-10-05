package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type zoneDeleteExecutor struct {
	calls   []executor.CommandSpec
	outputs map[string]string
	fail    string
	warning string
	codes   map[string]int
}

func (e *zoneDeleteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_zone.delete.")
	if stage == e.fail {
		return executor.CommandResult{}, errors.New("private diagnostic")
	}
	var diagnostic []byte
	if spec.Mutating || stage == e.warning {
		diagnostic = []byte("native notice")
	}
	return executor.CommandResult{Stdout: []byte(e.outputs[stage]), Stderr: diagnostic, ExitCode: e.codes[stage]}, nil
}
func zoneDeleteFixture(realm string) *zoneDeleteExecutor {
	zone := `{"id":"z","name":"secondary","realm_id":"` + realm + `","system_key":{"secret_key":"private"}}`
	group := `{"id":"g","name":"group","realm_id":"` + realm + `","master_zone":"primary","zones":[{"id":"primary","name":"primary","log_data":true},{"id":"z","name":"secondary","log_data":true}]}`
	after := `{"id":"g","name":"group","realm_id":"` + realm + `","master_zone":"primary","zones":[{"id":"primary","name":"primary","log_data":false}]}`
	period := `{"id":"new","realm_id":"r","epoch":2,"master_zone":"primary","master_zonegroup":"g","period_map":{"zonegroups":[` + after + `]}}`
	groups := `{"zonegroups":["group"],"default_info":"g"}`
	return &zoneDeleteExecutor{codes: map[string]int{"absence": 2}, outputs: map[string]string{
		"zones_before":  `{"zones":["secondary","primary"],"default_info":"primary"}`,
		"zones_recheck": `{"zones":["secondary","primary"],"default_info":"primary"}`,
		"zones_after":   `{"zones":["primary"],"default_info":"primary"}`,
		"identity":      zone, "identity_recheck": zone,
		"groups_before": groups, "groups_recheck": groups, "groups_after": groups,
		"group_before_0": group, "group_recheck_0": group, "group_after_0": after,
		"realm_before":     `{"id":"r","current_period":"old"}`,
		"period_before":    `{"id":"old","realm_id":"r","master_zone":"primary"}`,
		"period.pre_check": `{"id":"r","current_period":"old"}`,
		"period.commit":    period, "period.realm_post_check": `{"id":"r","current_period":"new"}`,
		"period.period_post_check": period, "published": period,
	}}
}
func TestZoneDeleteChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		request := Request{ClusterID: cluster, Action: "rgw_zone.delete", Parameters: map[string]any{"zone_id": "z", "name": "secondary", "realm_id": realm, "confirm_delete": true}}
		runner := zoneDeleteFixture(realm)
		s.executor = runner
		result, err := s.Execute(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		if result.Details.(map[string]any)["period_published"] != (realm != "") {
			t.Fatal("wrong period scope")
		}
		writes := [][]string{}
		for _, call := range runner.calls {
			if call.Binary != executor.BinaryRGWAdmin {
				t.Fatal("wrong binary")
			}
			if call.Mutating {
				writes = append(writes, call.Args)
			}
		}
		wanted := [][]string{{"zone", "delete", "--zone-id", "z"}}
		if realm != "" {
			wanted = append(wanted, []string{"period", "update", "--commit", "--realm-id", "r", "--format", "json"})
		}
		if !reflect.DeepEqual(writes, wanted) {
			t.Fatalf("unexpected writes %v", writes)
		}
		for _, call := range runner.calls {
			stage := strings.TrimPrefix(call.ID, "rgw_zone.delete.")
			for _, mode := range []string{"error", "exit"} {
				failed := zoneDeleteFixture(realm)
				if mode == "error" {
					failed.fail = stage
				} else {
					failed.codes[stage] = 5
				}
				s.executor = failed
				_, err := s.Execute(context.Background(), request)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe failure at %s/%s: %v", stage, mode, err)
				}
				if failed.calls[len(failed.calls)-1].ID != call.ID {
					t.Fatalf("continued past %s", stage)
				}
			}
		}
	}
}
func TestZoneDeleteRejectsDrift(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tc := range []struct{ stage, body string }{
		{"zones_before", `{"zones":["secondary","primary"],"default_info":"z"}`},
		{"identity", `{"id":"z","name":"changed","realm_id":"r"}`},
		{"group_before_0", `{"id":"g","name":"group","realm_id":"r","master_zone":"z","zones":[{"id":"z"},{"id":"primary"}]}`},
		{"group_before_0", `{"id":"g","name":"group","realm_id":"other","master_zone":"primary","zones":[{"id":"z"},{"id":"primary"}]}`},
		{"groups_recheck", `{"zonegroups":["group","new"]}`},
		{"identity_recheck", `{}`}, {"group_recheck_0", `{}`},
		{"zones_after", `{"zones":["secondary","primary"],"default_info":"primary"}`},
		{"groups_after", `{"zonegroups":[]}`}, {"group_after_0", `{}`}, {"published", `{}`},
	} {
		runner := zoneDeleteFixture("r")
		runner.outputs[tc.stage] = tc.body
		s.executor = runner
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zone.delete", Parameters: map[string]any{"zone_id": "z", "name": "secondary", "realm_id": "r", "confirm_delete": true}})
		if err == nil {
			t.Fatalf("accepted drift at %s", tc.stage)
		}
		if strings.Contains(tc.stage, "before") || strings.Contains(tc.stage, "recheck") || tc.stage == "identity" {
			for _, call := range runner.calls {
				if call.Mutating {
					t.Fatal("wrote despite drift")
				}
			}
		}
	}
	for _, p := range []map[string]any{nil, {"zone_id": "z", "name": "n", "confirm_delete": true}, {"zone_id": "z", "name": "n", "realm_id": "", "confirm_delete": false}} {
		if _, err := buildZoneDelete(p); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
}
func TestZoneDeleteExpectedNativeMembership(t *testing.T) {
	group := periodDocument([]byte(`{"realm_id":"r","master_zone":"a","zones":[{"id":"a","log_data":true},{"id":"b","log_data":true},{"id":"z","log_data":true}]}`))
	next, ok := zoneDeleteExpectedGroup(group, "z", "r")
	if !ok || len(next["zones"].([]any)) != 2 || next["zones"].([]any)[0].(map[string]any)["log_data"] != true {
		t.Fatal("native multi-member log_data mismatch")
	}
	if len(group["zones"].([]any)) != 3 {
		t.Fatal("mutated before snapshot")
	}
	for _, raw := range []string{`{}`, `{"zones":[{"id":"z"},{"id":"z"}]}`, `{"realm_id":"r","master_zone":"missing","zones":[{"id":"z"},{"id":"a"}]}`} {
		if _, ok := zoneDeleteExpectedGroup(periodDocument([]byte(raw)), "z", "r"); ok {
			t.Fatal("invalid membership accepted")
		}
	}
}

func TestZoneDeleteVerifiesAllGroupsAndPublishedMembers(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, scenario := range []string{"two_groups", "ignored_write", "wrong_log_data", "published_target", "published_missing_group", "default_warning", "default_drift"} {
		t.Run(scenario, func(t *testing.T) {
			runner := zoneDeleteFixture("r")
			unrelated := `{"id":"other","name":"other","realm_id":"outside","master_zone":"remote","zones":[{"id":"remote","log_data":false}]}`
			for _, stage := range []string{"groups_before", "groups_recheck", "groups_after"} {
				runner.outputs[stage] = `{"zonegroups":["group","other"],"default_info":"g"}`
			}
			for _, stage := range []string{"group_before_1", "group_recheck_1", "group_after_1"} {
				runner.outputs[stage] = unrelated
			}
			switch scenario {
			case "ignored_write":
				runner.outputs["group_after_0"] = runner.outputs["group_before_0"]
			case "wrong_log_data":
				runner.outputs["group_after_0"] = strings.ReplaceAll(runner.outputs["group_after_0"], `"log_data":false`, `"log_data":true`)
			case "published_target":
				runner.outputs["published"] = strings.Replace(runner.outputs["published"], runner.outputs["group_after_0"], runner.outputs["group_before_0"], 1)
			case "published_missing_group":
				runner.outputs["published"] = strings.Replace(runner.outputs["published"], runner.outputs["group_after_0"], unrelated, 1)
			case "default_warning":
				runner.warning = "zones_before"
			case "default_drift":
				runner.outputs["zones_recheck"] = `{"zones":["secondary","primary"],"default_info":"z"}`
			}
			s.executor = runner
			_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zone.delete", Parameters: map[string]any{"zone_id": "z", "name": "secondary", "realm_id": "r", "confirm_delete": true}})
			if (err == nil) != (scenario == "two_groups") {
				t.Fatalf("unexpected result: %v", err)
			}
			if scenario == "default_warning" || scenario == "default_drift" {
				for _, call := range runner.calls {
					if call.Mutating {
						t.Fatal("wrote with unknown or changed default")
					}
				}
			}
		})
	}
}
