package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestIngressListenerUpdate(t *testing.T) {
	current := `[{"service_name":"ingress.rgw.a","service_type":"ingress","service_id":"rgw.a","spec":{"backend_service":"rgw.a","virtual_ip":"192.0.2.1/24","frontend_port":443,"monitor_port":9000,"ssl":true,"ssl_key":"retained-key","keepalive_only":false,"future":18446744073709551615}}]`
	p := map[string]any{"service_type": "ingress", "virtual_ip": "2001:db8::10/64", "frontend_port": 8443, "monitor_port": 9001}
	s, _, id := newCephUserService(t)
	e := &directoryRenameExecutor{outputs: map[string]string{"service.update.pre_check": current, "service.update": "Scheduled ingress.rgw.a update..."}}
	s.executor = e
	if _, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "service.update", ResourceKey: "service/ingress.rgw.a", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	if len(e.specs) != 3 || e.specs[0].Mutating || !e.specs[1].Mutating {
		t.Fatal("wrong listener update chain")
	}
	var outer map[string]json.RawMessage
	var spec map[string]json.RawMessage
	json.Unmarshal(e.specs[1].Stdin, &outer)
	json.Unmarshal(outer["spec"], &spec)
	for key, want := range map[string]string{"virtual_ip": `"2001:db8::10/64"`, "frontend_port": "8443", "monitor_port": "9001", "ssl_key": `"retained-key"`, "backend_service": `"rgw.a"`, "future": "18446744073709551615"} {
		if string(spec[key]) != want {
			t.Fatalf("field %s changed incorrectly", key)
		}
	}
	cmd, err := build(Request{Action: "service.update", ResourceKey: "service/ingress.rgw.a"}, p)
	if err != nil {
		t.Fatal(err)
	}
	for _, replacement := range []string{`"keepalive_only":true`, `"keepalive_only":null`, `"keepalive_only":false,"virtual_ips_list":["192.0.2.1/24"]`} {
		bad := strings.Replace(current, `"keepalive_only":false`, replacement, 1)
		if _, err := mergeServiceSpec([]byte(bad), cmd.stdin, "ingress.rgw.a"); err == nil {
			t.Fatal("unsupported native mode accepted")
		}
	}
	for _, raw := range []string{`{"virtual_ip":"192.0.2.1/24"}`, `{"frontend_port":443,"monitor_port":9000}`, `{"virtual_ip":"invalid","frontend_port":443,"monitor_port":9000}`, `{"virtual_ip":"192.0.2.1/24","frontend_port":443,"monitor_port":443}`, `{"virtual_ip":"192.0.2.1/24","frontend_port":0,"monitor_port":9000}`} {
		var values map[string]any
		json.Unmarshal([]byte(raw), &values)
		if _, err := ingressListenerParameters(values); err == nil {
			t.Fatal("invalid listener patch accepted")
		}
	}
}
