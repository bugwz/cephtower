package mutation

import (
	"context"
	"encoding/json"
	"testing"
)

func TestIngressNetworkUpdate(t *testing.T) {
	for _, networks := range []any{[]string{}, []string{"192.0.2.0/24", "2001:db8::/64"}} {
		s, _, id := newCephUserService(t)
		e := &directoryRenameExecutor{outputs: map[string]string{"service.update.pre_check": `[{"service_name":"ingress.rgw.a","service_type":"ingress","service_id":"rgw.a","spec":{"virtual_interface_networks":["198.51.100.0/24"],"backend_service":"rgw.a","ssl":true,"ssl_key":"retained","future":18446744073709551615}}]`, "service.update": "Scheduled ingress.rgw.a update..."}}
		s.executor = e
		_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "service.update", ResourceKey: "service/ingress.rgw.a", Parameters: map[string]any{"service_type": "ingress", "virtual_interface_networks": networks}})
		if err != nil {
			t.Fatal(err)
		}
		if len(e.specs) != 3 || !e.specs[1].Mutating {
			t.Fatal("wrong update chain")
		}
		var outer map[string]json.RawMessage
		var spec map[string]json.RawMessage
		if err := json.Unmarshal(e.specs[1].Stdin, &outer); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(outer["spec"], &spec); err != nil {
			t.Fatal(err)
		}
		want, _ := json.Marshal(networks)
		for key, value := range map[string]string{"virtual_interface_networks": string(want), "ssl_key": `"retained"`, "backend_service": `"rgw.a"`, "future": "18446744073709551615"} {
			if string(spec[key]) != value {
				t.Fatalf("incorrect field %s", key)
			}
		}
	}
	for _, value := range []any{nil, "192.0.2.0/24", []string{"bad"}, []any{42}} {
		if _, err := ingressServiceSpec(map[string]any{"virtual_interface_networks": value}, "ingress", "service.update"); err == nil {
			t.Fatal("invalid network accepted")
		}
	}
}
