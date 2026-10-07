package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestIngressServiceCreation(t *testing.T) {
	parameters := func() map[string]any {
		return map[string]any{"service_type": "ingress", "service_id": "rgw.a", "backend_service": "rgw.a", "virtual_ip": "192.0.2.10/24", "frontend_port": 8080, "monitor_port": 9000}
	}
	for _, vip := range []string{"192.0.2.10/24", "2001:db8::10/64"} {
		p := parameters()
		p["virtual_ip"] = vip
		s, _, id := newCephUserService(t)
		e := &directoryRenameExecutor{outputs: map[string]string{"service.create": "Scheduled ingress.rgw.a update..."}}
		s.executor = e
		if _, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "service.create", ResourceKey: "service", Parameters: p}); err != nil {
			t.Fatal(err)
		}
		if len(e.specs) != 2 || !reflect.DeepEqual(e.specs[0].Args, []string{"orch", "apply", "-i", "-", "--no-overwrite"}) {
			t.Fatalf("wrong native chain: %+v", e.specs)
		}
		var spec map[string]any
		json.Unmarshal(e.specs[0].Stdin, &spec)
		detail := spec["spec"].(map[string]any)
		if spec["service_id"] != "rgw.a" || detail["backend_service"] != "rgw.a" || detail["virtual_ip"] != vip || detail["frontend_port"] != float64(8080) || detail["monitor_port"] != float64(9000) {
			t.Fatal(spec)
		}
	}
	for key, values := range map[string][]any{"service_id": {""}, "backend_service": {nil, "", "mon.a"}, "virtual_ip": {"192.0.2.10", "bad/24"}, "frontend_port": {nil, 0, -1, 65536, 1.5, true, 9000}, "monitor_port": {nil, 0, 65536}} {
		for _, value := range values {
			p := parameters()
			p[key] = value
			if _, err := build(Request{Action: "service.create"}, p); err == nil {
				t.Fatalf("accepted %s=%v", key, value)
			}
		}
	}
	p := parameters()
	if _, err := build(Request{Action: "service.update", ResourceKey: "service/ingress.rgw.a"}, p); err == nil {
		t.Fatal("unsupported listener changes silently accepted")
	}
	delete(p, "backend_service")
	delete(p, "virtual_ip")
	delete(p, "frontend_port")
	delete(p, "monitor_port")
	cmd, err := build(Request{Action: "service.update", ResourceKey: "service/ingress.rgw.a"}, p)
	if err != nil {
		t.Fatal(err)
	}
	merged, err := mergeServiceSpec([]byte(`[{"service_name":"ingress.rgw.a","service_type":"ingress","service_id":"rgw.a","spec":{"backend_service":"rgw.a","ssl":true,"ssl_cert":"preserved"}}]`), cmd.stdin, "ingress.rgw.a")
	var spec map[string]any
	json.Unmarshal(merged, &spec)
	if err != nil || spec["spec"].(map[string]any)["ssl_cert"] != "preserved" {
		t.Fatalf("lost ingress settings: %s %v", merged, err)
	}
}
