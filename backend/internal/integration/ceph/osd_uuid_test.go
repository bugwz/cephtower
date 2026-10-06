package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestOSDUUIDUsesMatchingNativeRecord(t *testing.T) {
	for _, raw := range []string{`null`, `""`, `"00000000-0000-0000-0000-000000000000"`, `"12345678-1234-1234-1234-123456789abc"`, `1`, `true`, `{}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.osd_tree": []byte(`{"nodes":[{"id":0,"name":"osd.0","type":"osd"},{"id":2,"name":"osd.2","type":"osd"}]}`),
			"collect.osd_dump": []byte(`{"osds":[{"osd":1,"uuid":"other"},{"osd":0,"uuid":` + raw + `}]}`),
		}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
		if raw == `1` || raw == `true` || raw == `{}` {
			if err == nil || len(rows) != 0 {
				t.Fatalf("invalid uuid accepted: %s", raw)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, row := range rows {
			if row.Kind != "osd" {
				continue
			}
			count++
			got := row.Payload.(cephdomain.OSD)
			encoded, err := json.Marshal(got.UUID)
			if err != nil {
				t.Fatal(err)
			}
			want := raw
			if got.ID == 2 {
				want = `null`
			}
			if string(encoded) != want {
				t.Fatalf("osd %d uuid %s != %s", got.ID, encoded, want)
			}
		}
		if count != 2 {
			t.Fatalf("OSDs=%d", count)
		}
	}
}
