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

type realmDeleteExecutor struct {
	specs   []executor.CommandSpec
	outputs map[string]string
	codes   map[string]int
	fail    string
}

func (e *realmDeleteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	stage := strings.TrimPrefix(spec.ID, "rgw_realm.delete.")
	if stage == e.fail {
		return executor.CommandResult{}, errors.New("private diagnostic")
	}
	return executor.CommandResult{Stdout: []byte(e.outputs[stage]), ExitCode: e.codes[stage]}, nil
}
func newRealmDeleteExecutor() *realmDeleteExecutor {
	return &realmDeleteExecutor{outputs: map[string]string{
		"list_before": `{"default_info":"other-id","realms":["other","realm"]}`,
		"identity":    `{"id":"realm-id","name":"realm","current_period":"period"}`,
		"list_after":  `{"default_info":"other-id","realms":["other"]}`,
	}, codes: map[string]int{"absence": 2}}
}
func realmDeleteParameters() map[string]any {
	return map[string]any{"realm_id": "realm-id", "name": "realm", "expected_current_period": "period", "confirm_delete": true}
}
func TestRealmDeleteExecution(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	runner := newRealmDeleteExecutor()
	s.executor = runner
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.delete", Parameters: realmDeleteParameters()})
	if err != nil || result.Details.(map[string]any)["realm_absence_verified"] != true {
		t.Fatalf("deletion failed: %v", err)
	}
	if !Supports("rgw_realm.delete") || len(runner.specs) != 5 {
		t.Fatal("wrong dispatch")
	}
	writes := 0
	for _, spec := range runner.specs {
		if spec.Binary != executor.BinaryRGWAdmin || spec.Timeout <= 0 {
			t.Fatal("wrong command")
		}
		if spec.Mutating {
			writes++
			if !reflect.DeepEqual(spec.Args, []string{"realm", "rm", "--realm-id", "realm-id"}) {
				t.Fatal("wrong deletion scope")
			}
		}
	}
	if writes != 1 {
		t.Fatal("unexpected write count")
	}
	runner = newRealmDeleteExecutor()
	runner.outputs["list_before"] = `{"default_info":"","realms":["realm"]}`
	runner.outputs["list_after"] = `{"default_info":"","realms":[]}`
	s.executor = runner
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.delete", Parameters: realmDeleteParameters()}); err != nil {
		t.Fatalf("last non-default realm deletion failed: %v", err)
	}
	for _, stage := range []string{"list_before", "identity", "delete", "absence", "list_after"} {
		runner = newRealmDeleteExecutor()
		runner.fail = stage
		s.executor = runner
		_, err = s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.delete", Parameters: realmDeleteParameters()})
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Retryable || strings.Contains(err.Error(), "private diagnostic") {
			t.Fatalf("unsafe failure at %s", stage)
		}
		if strings.TrimPrefix(runner.specs[len(runner.specs)-1].ID, "rgw_realm.delete.") != stage {
			t.Fatal("continued after failure")
		}
	}
}
func TestRealmDeleteRejectsChangedEvidence(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tc := range []struct{ stage, raw string }{
		{"list_before", `{"default_info":"realm-id","realms":["realm"]}`},
		{"list_before", `{"realms":["realm"]}`},
		{"list_before", `{"default_info":"other","realms":["realm","realm"]}`},
		{"list_before", `{"default_info":"other","realms":[]}`},
		{"identity", `{"id":"wrong","name":"realm","current_period":"period"}`},
		{"identity", `{"id":"realm-id","name":"renamed","current_period":"period"}`},
		{"identity", `{"id":"realm-id","name":"realm","current_period":"changed"}`},
		{"list_after", `{"default_info":"other-id","realms":["other","realm"]}`},
		{"list_after", `{"default_info":"changed","realms":["other"]}`},
		{"list_after", `{"default_info":"other-id","realms":[]}`},
	} {
		runner := newRealmDeleteExecutor()
		runner.outputs[tc.stage] = tc.raw
		s.executor = runner
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.delete", Parameters: realmDeleteParameters()}); err == nil {
			t.Fatalf("accepted %s", tc.raw)
		}
		if tc.stage != "list_after" {
			for _, spec := range runner.specs {
				if spec.Mutating {
					t.Fatal("wrote with stale evidence")
				}
			}
		}
	}
	for _, code := range []int{0, 1, 13, 22} {
		runner := newRealmDeleteExecutor()
		runner.codes["absence"] = code
		s.executor = runner
		if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.delete", Parameters: realmDeleteParameters()}); err == nil {
			t.Fatal("absence inferred from wrong exit code")
		}
	}
	runner := newRealmDeleteExecutor()
	runner.codes["delete"] = 5
	s.executor = runner
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.delete", Parameters: realmDeleteParameters()}); err == nil || len(runner.specs) != 3 {
		t.Fatal("nonzero deletion exit continued")
	}
	for _, field := range []string{"realm_id", "name", "expected_current_period", "confirm_delete"} {
		p := realmDeleteParameters()
		delete(p, field)
		if _, err := buildRealmDelete(p); err == nil {
			t.Fatal("incomplete request accepted")
		}
	}
}
