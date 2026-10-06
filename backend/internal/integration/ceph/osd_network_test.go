package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestOSDNetworkAddresses(t *testing.T) {
	for _, field := range []string{"public_addrs", "cluster_addrs", "heartbeat_front_addrs", "heartbeat_back_addrs"} {
		for _, raw := range []string{`null`, `{"addrvec":[]}`, `{"addrvec":[{"type":"v2","addr":"[::1]:3300","nonce":4294967295},{"type":"v1","addr":"host:6800","nonce":0}]}`} {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.osd_tree": []byte(`{"nodes":[{"id":0,"name":"osd.0","type":"osd"}]}`),
				"collect.osd_dump": []byte(`{"osds":[{"osd":0,"` + field + `":` + raw + `}]}`),
			}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, row := range rows {
				if row.Kind != "osd" {
					continue
				}
				found = true
				encoded, err := json.Marshal(row.Payload.(cephdomain.OSD))
				if err != nil {
					t.Fatal(err)
				}
				var got map[string]json.RawMessage
				if err := json.Unmarshal(encoded, &got); err != nil {
					t.Fatal(err)
				}
				if string(got[field]) != raw {
					t.Fatalf("%s=%s want %s", field, got[field], raw)
				}
			}
			if !found {
				t.Fatal("OSD missing")
			}
		}
	}
}
