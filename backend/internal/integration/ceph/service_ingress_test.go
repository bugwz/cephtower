package ceph

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestIngressSettingsProjection(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"backend_service":"rgw.a","virtual_ip":"192.0.2.10/24","frontend_port":443,"monitor_port":9000,"ssl":false,"keepalive_only":true,"virtual_interface_networks":["192.0.2.0/24","2001:db8::/64"],"ssl_key":"secret-key-fixture","ssl_cert":"secret-cert-fixture","monitor_password":"secret-password-fixture","unknown_sensitive":"secret-future-fixture"}`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.service": []byte(`[{"service_name":"ingress.rgw.a","service_type":"ingress","spec":` + raw + `}]`)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "service" {
				continue
			}
			found = true
			encoded, _ := json.Marshal(row.Payload)
			if strings.Contains(string(encoded), "secret-") || strings.Contains(string(encoded), "ssl_key") {
				t.Fatal("secret ingress configuration reached inventory")
			}
			var payload map[string]any
			json.Unmarshal(encoded, &payload)
			if raw == `null` {
				if payload["ingress"] != nil {
					t.Fatal("null settings became known")
				}
				continue
			}
			ingress := payload["ingress"].(map[string]any)
			if raw == `{}` {
				if ingress["ssl"] != nil || ingress["frontend_port"] != nil {
					t.Fatal("unknown configuration became false or zero")
				}
				continue
			}
			if ingress["ssl"] != false || ingress["keepalive_only"] != true || ingress["frontend_port"] != float64(443) || ingress["backend_service"] != "rgw.a" || len(ingress["virtual_interface_networks"].([]any)) != 2 {
				t.Fatal("native ingress configuration lost")
			}
		}
		if !found {
			t.Fatal("no service observation")
		}
	}
	for _, raw := range []string{`[]`, `{"ssl":"false"}`, `{"frontend_port":0}`, `{"monitor_port":65536}`, `{"virtual_interface_networks":[1]}`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.service": []byte(`[{"service_name":"ingress.rgw.a","service_type":"ingress","spec":` + raw + `}]`)}}}
		if _, err := provider.Collect(context.Background(), ClusterAccess{}, "topology"); err == nil {
			t.Fatal("malformed ingress configuration accepted")
		}
	}
}
