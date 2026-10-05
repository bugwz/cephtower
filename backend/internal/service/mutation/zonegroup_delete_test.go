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

type zonegroupDeleteExecutor struct {
	calls   []executor.CommandSpec
	outputs map[string]string
	fail    string
	codes   map[string]int
}

func (e *zonegroupDeleteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_zonegroup.delete.")
	if stage == e.fail {
		return executor.CommandResult{}, errors.New("private diagnostic")
	}
	return executor.CommandResult{Stdout: []byte(e.outputs[stage]), ExitCode: e.codes[stage]}, nil
}
func zonegroupDeleteFixture(realm string) *zonegroupDeleteExecutor {
	group := `{"id":"g","name":"secondary","realm_id":"` + realm + `","is_master":false,"zones":[{"id":"z","name":"zone"}]}`
	period := `{"id":"p","realm_id":"r","epoch":2,"master_zonegroup":"master","period_map":{"zonegroups":[{"id":"master"}]}}`
	return &zonegroupDeleteExecutor{codes: map[string]int{"absence": 2}, outputs: map[string]string{
		"list_before": `{"zonegroups":["secondary","primary"],"default_info":"master"}`,
		"identity":    group, "identity_recheck": group,
		"realm_before":     `{"id":"r","current_period":"old"}`,
		"period_before":    `{"id":"old","realm_id":"r","master_zonegroup":"master"}`,
		"master_before":    `{"id":"master","realm_id":"r","is_master":true}`,
		"list_after":       `{"zonegroups":["primary"],"default_info":"master"}`,
		"period.pre_check": `{"id":"r","current_period":"old"}`,
		"period.commit":    period, "period.period_post_check": period, "published_absence": period,
		"period.realm_post_check": `{"id":"r","current_period":"p"}`,
	}}
}
func zonegroupDeleteParams(realm string) map[string]any {
	return map[string]any{"zonegroup_id": "g", "name": "secondary", "realm_id": realm, "expected_zones": []any{"z"}, "confirm_delete": true}
}
func TestZonegroupDeletePreservesZonesAndPublishesPeriod(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		runner := zonegroupDeleteFixture(realm)
		s.executor = runner
		result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.delete", Parameters: zonegroupDeleteParams(realm)})
		if err != nil {
			t.Fatal(err)
		}
		if result.Details.(map[string]any)["period_published"] != (realm != "") {
			t.Fatal("wrong publication result")
		}
		writes := [][]string{}
		for _, spec := range runner.calls {
			if spec.Binary != executor.BinaryRGWAdmin {
				t.Fatal("unexpected binary")
			}
			if spec.Mutating {
				writes = append(writes, spec.Args)
			}
		}
		wanted := [][]string{{"zonegroup", "delete", "--zonegroup-id", "g"}}
		if realm != "" {
			wanted = append(wanted, []string{"period", "update", "--commit", "--realm-id", "r", "--format", "json"})
		}
		if !reflect.DeepEqual(writes, wanted) {
			t.Fatalf("unexpected writes: %v", writes)
		}
		for _, spec := range runner.calls {
			stage := strings.TrimPrefix(spec.ID, "rgw_zonegroup.delete.")
			failed := zonegroupDeleteFixture(realm)
			failed.fail = stage
			s.executor = failed
			_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.delete", Parameters: zonegroupDeleteParams(realm)})
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Retryable || strings.Contains(err.Error(), "private diagnostic") {
				t.Fatalf("unsafe error %s: %v", stage, err)
			}
			if failed.calls[len(failed.calls)-1].ID != spec.ID {
				t.Fatalf("continued after %s", stage)
			}
		}
	}
}
func TestZonegroupDeleteRejectsDrift(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tc := range []struct{ stage, raw string }{
		{"list_before", `{"zonegroups":["secondary"],"default_info":"g"}`},
		{"identity", `{"id":"g","name":"secondary","realm_id":"r","is_master":true,"zones":[]}`},
		{"identity", `{"id":"g","name":"secondary","realm_id":"r","is_master":false,"zones":[{"id":"changed"}]}`},
		{"identity_recheck", `{}`}, {"master_before", `{}`}, {"period_before", `{"id":"old","realm_id":"r","master_zonegroup":"g"}`},
		{"list_after", `{"zonegroups":["primary","secondary"],"default_info":"master"}`},
		{"published_absence", `{"id":"p","realm_id":"r","master_zonegroup":"master","period_map":{"zonegroups":[{"id":"master"},{"id":"g"}]}}`},
	} {
		runner := zonegroupDeleteFixture("r")
		runner.outputs[tc.stage] = tc.raw
		s.executor = runner
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.delete", Parameters: zonegroupDeleteParams("r")}); err == nil {
			t.Fatalf("accepted %s", tc.stage)
		}
		if tc.stage != "list_after" && tc.stage != "published_absence" {
			for _, spec := range runner.calls {
				if spec.Mutating {
					t.Fatal("wrote with stale evidence")
				}
			}
		}
	}
	for _, field := range []string{"zonegroup_id", "realm_id", "expected_zones", "confirm_delete"} {
		p := zonegroupDeleteParams("r")
		delete(p, field)
		if _, err := buildZonegroupDelete(p); err == nil {
			t.Fatal("incomplete request accepted")
		}
	}
	for _, raw := range []string{`{}`, `{"id":"p","realm_id":"r","master_zonegroup":"master","period_map":{"zonegroups":[]}}`, `{"id":"p","realm_id":"r","master_zonegroup":"master","period_map":{"zonegroups":[{"id":"master"},{"id":"master"}]}}`} {
		if zonegroupAbsentFromPeriod(periodDocument([]byte(raw)), "r", "g") {
			t.Fatal("invalid publication accepted")
		}
	}
}
