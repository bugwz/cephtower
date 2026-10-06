package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestOSDHistoryEpochs(t *testing.T) {
	for _, field := range []string{"last_clean_begin", "last_clean_end", "up_from", "up_thru", "down_at", "lost_at"} {
		for _, raw := range []string{`null`, `0`, `4294967295`, `-1`, `4294967296`, `0.5`} {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.osd_tree": []byte(`{"nodes":[{"id":0,"name":"osd.0","type":"osd"}]}`),
				"collect.osd_dump": []byte(`{"osds":[{"osd":0,"` + field + `":` + raw + `}]}`),
			}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
			if raw == `-1` || raw == `4294967296` || raw == `0.5` {
				if err == nil {
					t.Fatal("invalid epoch accepted")
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
				encoded, err := json.Marshal(row.Payload.(cephdomain.OSD))
				if err != nil {
					t.Fatal(err)
				}
				var got map[string]json.RawMessage
				if err := json.Unmarshal(encoded, &got); err != nil {
					t.Fatal(err)
				}
				if string(got[field]) != raw {
					t.Fatalf("%s=%s, want %s", field, got[field], raw)
				}
			}
			if !found {
				t.Fatal("OSD missing")
			}
		}
	}
}
