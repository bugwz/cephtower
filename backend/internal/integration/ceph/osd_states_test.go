package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestOSDStatesRetainUnknown(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		want    string
		invalid bool
	}{
		{`[]`, `null`, false},
		{`[{"osd":0}]`, `null`, false},
		{`[{"osd":0,"up":null,"in":null}]`, `null`, false},
		{`[{"osd":0,"up":0,"in":0}]`, `false`, false},
		{`[{"osd":0,"up":1,"in":1}]`, `true`, false},
		{`[{"osd":0,"up":2}]`, ``, true},
		{`[{"osd":0,"in":-1}]`, ``, true},
		{`[{"osd":0,"up":true}]`, ``, true},
	} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.osd_tree": []byte(`{"nodes":[{"id":0,"name":"osd.0","type":"osd"}]}`),
			"collect.osd_dump": []byte(`{"osds":` + tc.raw + `}`),
		}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
		if tc.invalid {
			if err == nil || len(rows) != 0 {
				t.Fatalf("accepted %s", tc.raw)
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
			for _, state := range []*bool{got.Up, got.In} {
				encoded, err := json.Marshal(state)
				if err != nil {
					t.Fatal(err)
				}
				if string(encoded) != tc.want {
					t.Fatalf("%s: state %s != %s", tc.raw, encoded, tc.want)
				}
			}
		}
		if !found {
			t.Fatal("OSD missing")
		}
	}
}

func TestOSDDumpRejectsAmbiguousIdentity(t *testing.T) {
	for _, raw := range []string{`{}`, `{"osds":null}`, `{"osds":[null]}`, `{"osds":[{}]}`, `{"osds":[{"osd":null}]}`, `{"osds":[{"osd":-1}]}`, `{"osds":[{"osd":0},{"osd":0}]}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_dump": []byte(raw)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
		if err == nil || len(rows) != 0 {
			t.Fatalf("accepted ambiguous osd dump %s: %v", raw, err)
		}
	}
}
