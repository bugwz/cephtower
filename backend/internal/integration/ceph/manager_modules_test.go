package ceph

import (
	"context"
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
