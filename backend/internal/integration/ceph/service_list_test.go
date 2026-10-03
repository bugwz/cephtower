package ceph

import (
	"context"
	"encoding/json"
	"testing"
)

func TestServiceRuntimeMetadataReachesPayload(t *testing.T) {
	for _, status := range []string{`{}`, `{"container_image_name":"quay.io/ceph/ceph:v20","container_image_id":"sha256:abc","service_url":"https://node.example:8443/","virtual_ip":"192.0.2.10","created":"2026-10-03T00:00:00Z"}`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.service": []byte(`[{"service_name":"mgr","service_type":"mgr","status":` + status + `}]`)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		var expected map[string]any
		if err := json.Unmarshal([]byte(status), &expected); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "service" {
				continue
			}
			found = true
			data, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatal(err)
			}
			for native, exposed := range map[string]string{"container_image_name": "container_image_name", "container_image_id": "container_image_id", "service_url": "service_url", "virtual_ip": "virtual_ip", "created": "ceph_created_at"} {
				got, exists := payload[exposed]
				if !exists || got != expected[native] {
					t.Fatalf("%s: got %#v, want %#v", exposed, got, expected[native])
				}
			}
			if _, exists := payload["created_at"]; exists {
				t.Fatal("native creation time collides with cache metadata")
			}
		}
		if !found {
			t.Fatal("service missing")
		}
	}
}

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
