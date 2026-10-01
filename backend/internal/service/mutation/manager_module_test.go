package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestManagerModuleStateMatches(t *testing.T) {
	const active = `{"always_on_modules":[],"enabled_modules":["prometheus"],"force_disabled_modules":[],"disabled_modules":[]}`
	const inactive = `{"always_on_modules":[],"enabled_modules":[],"force_disabled_modules":[],"disabled_modules":[{"name":"prometheus","can_run":true}]}`
	const always = `{"always_on_modules":["prometheus"],"enabled_modules":[],"force_disabled_modules":[],"disabled_modules":[]}`
	const forced = `{"always_on_modules":["prometheus"],"enabled_modules":[],"force_disabled_modules":["prometheus"],"disabled_modules":[]}`
	for _, test := range []struct {
		data    string
		enabled bool
	}{{active, true}, {inactive, false}, {always, true}, {forced, false}} {
		if !managerModuleStateMatches([]byte(test.data), "prometheus", test.enabled) || managerModuleStateMatches([]byte(test.data), "prometheus", !test.enabled) {
			t.Fatalf("incorrect state: %s", test.data)
		}
		for _, enabled := range []bool{true, false} {
			if managerModuleStateMatches([]byte(test.data), "missing", enabled) {
				t.Fatal("unknown module was treated as known")
			}
		}
	}
	for _, data := range []string{
		`null`, `{}`, `[]`, active + `{}`,
		strings.Replace(active, `"enabled_modules":["prometheus"]`, `"enabled_modules":[null]`, 1),
		strings.Replace(active, `"enabled_modules":["prometheus"]`, `"enabled_modules":["prometheus","prometheus"]`, 1),
		strings.Replace(active, `"enabled_modules":["prometheus"]`, `"enabled_modules":[" prometheus"]`, 1),
		strings.Replace(active, `"disabled_modules":[]`, `"disabled_modules":null`, 1),
		strings.Replace(active, `"disabled_modules":[]`, `"disabled_modules":[{}]`, 1),
		strings.Replace(active, `"disabled_modules":[]`, `"disabled_modules":[{"name":"prometheus"}]`, 1),
		strings.Replace(active, `"always_on_modules":[]`, `"always_on_modules":["prometheus"]`, 1),
		strings.Replace(active, `"force_disabled_modules":[]`, `"force_disabled_modules":["prometheus"]`, 1),
	} {
		for _, enabled := range []bool{true, false} {
			if managerModuleStateMatches([]byte(data), "prometheus", enabled) {
				t.Fatalf("invalid state accepted: %s", data)
			}
		}
	}
}

func TestManagerModuleMutationConfirmsActivation(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, enabled := range []bool{true, false} {
		verb, state := "disable", `{"always_on_modules":[],"enabled_modules":[],"force_disabled_modules":[],"disabled_modules":[{"name":"prometheus"}]}`
		if enabled {
			verb, state = "enable", `{"always_on_modules":[],"enabled_modules":["prometheus"],"force_disabled_modules":[],"disabled_modules":[]}`
		}
		for _, failure := range []string{"", "invalid", "manager_module.update.post_check", "manager_module.update"} {
			e := &directoryRenameExecutor{outputs: map[string]string{"manager_module.update.post_check": state}, failID: failure}
			if failure == "invalid" {
				e.outputs["manager_module.update.post_check"] = `{}`
			}
			s.executor = e
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "manager_module.update", ResourceKey: "manager-module/prometheus", Parameters: map[string]any{"enabled": enabled}})
			if failure == "" && err != nil {
				t.Fatal(err)
			}
			if failure != "" && err == nil {
				t.Fatal("unconfirmed mutation succeeded")
			}
			if failure == "invalid" || failure == "manager_module.update.post_check" {
				var actionError *cephdomain.ActionError
				if !errors.As(err, &actionError) || actionError.Code != "post_check_failed" || actionError.Retryable {
					t.Fatalf("unsafe retry classification: %v", err)
				}
			}
			if !reflect.DeepEqual(e.specs[0].Args, []string{"mgr", "module", verb, "prometheus"}) || !e.specs[0].Mutating {
				t.Fatalf("wrong mutation: %+v", e.specs[0])
			}
			if failure == "manager_module.update" {
				if len(e.specs) != 1 {
					t.Fatal("post-check after failed write")
				}
				continue
			}
			if len(e.specs) != 2 || !reflect.DeepEqual(e.specs[1].Args, []string{"mgr", "module", "ls", "--format", "json"}) || e.specs[1].Mutating {
				t.Fatalf("wrong post-check: %+v", e.specs)
			}
		}
	}
}
