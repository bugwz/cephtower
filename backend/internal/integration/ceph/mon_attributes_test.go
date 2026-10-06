package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestMonitorPriorityAndWeight(t *testing.T) {
	for _, fields := range []string{``, `,"priority":null,"weight":null`, `,"priority":0,"weight":65535`, `,"priority":65535,"weight":0`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mon": []byte(`{"mons":[{"name":"ceph-node-1","rank":0` + fields + `}]}`)}}}
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
			encoded, err := json.Marshal(mon)
			if err != nil {
				t.Fatal(err)
			}
			var value map[string]any
			if err := json.Unmarshal(encoded, &value); err != nil {
				t.Fatal(err)
			}
			if fields == `` || fields == `,"priority":null,"weight":null` {
				if mon.Priority != nil || mon.Weight != nil || value["priority"] != nil || value["weight"] != nil {
					t.Fatalf("invented attributes: %s", encoded)
				}
			} else {
				if mon.Priority == nil || mon.Weight == nil || uint32(*mon.Priority)+uint32(*mon.Weight) != 65535 {
					t.Fatalf("lost attributes: %s", encoded)
				}
				if fields == `,"priority":0,"weight":65535` && *mon.Priority != 0 {
					t.Fatalf("lost zero priority: %s", encoded)
				}
				if fields == `,"priority":65535,"weight":0` && *mon.Weight != 0 {
					t.Fatalf("lost zero weight: %s", encoded)
				}
			}
		}
		if !found {
			t.Fatal("monitor missing")
		}
	}
	for _, field := range []string{"priority", "weight"} {
		for _, bad := range []string{`-1`, `65536`, `1.5`, `"1"`, `false`} {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mon": []byte(`{"mons":[{"name":"ceph-node-1","` + field + `":` + bad + `}]}`)}}}
			if _, err := p.Collect(context.Background(), ClusterAccess{}, "topology"); err == nil {
				t.Fatalf("accepted %s=%s", field, bad)
			}
		}
	}
}

func TestMonitorCrushLocation(t *testing.T) {
	for _, raw := range []string{`null`, `""`, `"[datacenter=dc1,rack=rack1]"`, `{}`, `[]`, `false`, `1`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mon": []byte(`{"mons":[{"name":"ceph-node-1","rank":0,"crush_location":` + raw + `}]}`)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		var expected *string
		valid := json.Unmarshal([]byte(raw), &expected) == nil
		if !valid {
			if err == nil {
				t.Fatalf("accepted invalid location %s", raw)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "mon" {
				continue
			}
			found = true
			got := row.Payload.(cephdomain.Monitor).CrushLocation
			if (got == nil) != (expected == nil) || (got != nil && *got != *expected) {
				t.Fatalf("location lost: %v expected %s", got, raw)
			}
			encoded, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["crush_location"]) != raw {
				t.Fatalf("location JSON changed: %s", encoded)
			}
		}
		if !found {
			t.Fatal("monitor missing")
		}
	}
}
