package ceph

import (
	"context"
	"testing"
)

func TestServiceListNativeEmptyResponse(t *testing.T) {
	for _, output := range []string{"No services reported", "No services reported\n", "[]"} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.service": []byte(output)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatalf("empty service response failed: %q: %v", output, err)
		}
		if len(rows) == 0 {
			t.Fatal("empty services discarded other topology observations")
		}
		for _, row := range rows {
			if row.Kind == "service" {
				t.Fatalf("unexpected service: %+v", row)
			}
		}
	}
}

func TestServiceListRejectsUnconfirmedInventory(t *testing.T) {
	for _, output := range []string{"", "null", "{}", "[null]", "[] {}", "No services reported: unavailable", `"No services reported"`, `[{"service_name":"mgr","service_type":"mgr"},{"service_name":"mgr","service_type":"mgr"}]`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.service": []byte(output)}}}
		if rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology"); err == nil || len(rows) != 0 {
			t.Fatalf("unconfirmed inventory accepted: %q", output)
		}
	}
}
