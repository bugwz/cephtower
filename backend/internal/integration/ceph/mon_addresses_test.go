package ceph

import (
	"context"
	"encoding/json"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestMonitorCollectionPreservesProtocolAddresses(t *testing.T) {
	for _, addresses := range []string{`[{"type":"v2","addr":"[2001:db8::1]:3300","nonce":0},{"type":"v1","addr":"192.0.2.1:6789","nonce":0}]`, `[]`, `null`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.mon": []byte(`{"mons":[{"name":"ceph-node-1","rank":0,"public_addrs":{"addrvec":` + addresses + `}}]}`),
		}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "mon" {
				continue
			}
			found = true
			mon := row.Payload.(cephdomain.Monitor)
			if addresses == `null` && mon.PublicAddresses != nil {
				t.Fatal("invented addresses")
			}
			if addresses == `[]` && (mon.PublicAddresses == nil || len(mon.PublicAddresses) != 0) {
				t.Fatal("lost empty list")
			}
			if addresses != `null` && addresses != `[]` {
				if len(mon.PublicAddresses) != 2 || mon.PublicAddresses[0].Type != "v2" || mon.PublicAddresses[0].Address != "[2001:db8::1]:3300" || mon.PublicAddresses[1].Type != "v1" || mon.PublicAddresses[1].Address != "192.0.2.1:6789" {
					t.Fatalf("lost native addresses: %+v", mon)
				}
			}
			raw, _ := json.Marshal(mon)
			var payload map[string]any
			if json.Unmarshal(raw, &payload) != nil {
				t.Fatal("invalid payload")
			}
			if _, present := payload["public_addresses"]; !present {
				t.Fatal("missing API field")
			}
		}
		if !found {
			t.Fatal("missing monitor")
		}
	}
}
