package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestOSDPrimaryAffinity(t *testing.T) {
	for _, raw := range []string{`null`, `0`, `0.25`, `1`, `-1`, `2`, `"1"`, `false`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.osd_tree": []byte(`{"nodes":[{"id":0,"name":"osd.0","type":"osd"}]}`),
			"collect.osd_dump": []byte(`{"osds":[{"osd":0,"primary_affinity":` + raw + `}]}`),
		}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
		if raw == `-1` || raw == `2` || raw == `"1"` || raw == `false` {
			if err == nil || len(rows) != 0 {
				t.Fatalf("accepted invalid affinity %s", raw)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "osd" {
				continue
			}
			found = true
			got := row.Payload.(cephdomain.OSD)
			encoded, err := json.Marshal(got.PrimaryAffinity)
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != raw {
				t.Fatalf("affinity %s != %s", encoded, raw)
			}
			if got.Reweight != nil {
				t.Fatal("affinity used as reweight")
			}
		}
		if !found {
			t.Fatal("OSD missing")
		}
	}
}
