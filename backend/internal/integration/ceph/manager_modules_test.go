package ceph

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestManagerModulesJoinActivationAndMetadata(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{override: map[string][]byte{
		"collect.mgr_module":          []byte(`{"enabled_modules":["dashboard"],"always_on_modules":["balancer"],"force_disabled_modules":["balancer"],"disabled_modules":[{"name":"prometheus"}]}`),
		"collect.mgr_module_metadata": []byte(`{"available_modules":[{"name":"dashboard","can_run":true,"module_options":{"port":{"type":"int","default_value":8080}}},{"name":"balancer","can_run":true},{"name":"prometheus","can_run":false,"error_string":"missing dependency"},{"name":"selftest"}]}`),
	}}}
	rows := p.collectManagerModules(context.Background(), ClusterAccess{}, time.Now())
	if len(rows) != 3 {
		t.Fatalf("rows=%+v", rows)
	}
	for _, row := range rows {
		value := row.Payload.(map[string]any)
		switch row.Name {
		case "dashboard":
			if value["enabled"] != true || value["options"] == nil {
				t.Fatalf("enabled module metadata lost: %+v", value)
			}
		case "balancer":
			if value["enabled"] != false || value["always_on"] != true || value["force_disabled"] != true {
				t.Fatalf("forced state lost: %+v", value)
			}
		case "prometheus":
			if value["enabled"] != false || value["can_run"] != false || value["error_string"] != "missing dependency" {
				t.Fatalf("disabled module lost: %+v", value)
			}
		}
	}
}

func TestManagerModulesRejectMalformedIdentityLists(t *testing.T) {
	for _, field := range []string{"enabled_modules", "always_on_modules", "force_disabled_modules", "disabled_modules", "available_modules"} {
		for _, raw := range []string{`[null]`, `[1]`, `[""]`, `[" a"]`, `["a","a"]`, `[{"name":"a"},null]`, `[{"name":"a"},{"name":"a"}]`, `[{"name":1}]`} {
			states := map[string]any{"enabled_modules": []any{}, "always_on_modules": []any{}, "force_disabled_modules": []any{}, "disabled_modules": []any{}}
			metadata := map[string]any{"available_modules": []any{map[string]any{"name": "dashboard"}}}
			var invalid any
			if err := json.Unmarshal([]byte(raw), &invalid); err != nil {
				t.Fatal(err)
			}
			if field == "available_modules" {
				metadata[field] = invalid
			} else {
				states[field] = invalid
			}
			stateJSON, _ := json.Marshal(states)
			metaJSON, _ := json.Marshal(metadata)
			p := NativeProvider{Executor: malformedExecutor{override: map[string][]byte{"collect.mgr_module": stateJSON, "collect.mgr_module_metadata": metaJSON}}}
			if rows := p.collectManagerModules(context.Background(), ClusterAccess{}, time.Now()); len(rows) != 0 {
				t.Fatalf("accepted %s=%s: %v", field, raw, rows)
			}
		}
	}
}
