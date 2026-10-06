package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestManagerInventoryRejectsAmbiguousIdentities(t *testing.T) {
	for _, raw := range []string{
		`{}`, `{"active_name":null,"standbys":[]}`, `{"active_name":"a"}`,
		`{"active_name":"a","standbys":null}`, `{"active_name":" a","standbys":[]}`,
		`{"active_name":"a","standbys":[null]}`, `{"active_name":"a","standbys":[{}]}`,
		`{"active_name":"a","standbys":[{"name":"a"}]}`,
		`{"active_name":"a","standbys":[{"name":"b"},{"name":"b"}]}`,
		`{"active_name":"a","standbys":[{"name":" b"}]}`,
	} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mgr": []byte(raw)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err == nil || len(rows) != 0 {
			t.Fatalf("accepted ambiguous inventory %s: %v", raw, err)
		}
	}
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mgr": []byte(`{"active_name":"","standbys":[]}`)}}}
	rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Kind == "mgr" {
			t.Fatal("invented manager for explicit empty inventory")
		}
	}
}

func TestManagerGIDPreservesNativePrecision(t *testing.T) {
	for _, raw := range []string{`null`, `0`, `9007199254740993`, `18446744073709551615`, `-1`, `1.5`, `"123"`, `true`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mgr": []byte(`{"active_name":"a","active_gid":` + raw + `,"standbys":[{"name":"b","gid":` + raw + `}]}`)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if raw == `-1` || raw == `1.5` || raw == `"123"` || raw == `true` {
			if err == nil || len(rows) != 0 {
				t.Fatalf("invalid gid accepted: %s", raw)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, row := range rows {
			if row.Kind != "mgr" {
				continue
			}
			count++
			encoded, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			var want any
			if raw != `null` {
				want = raw
			}
			if got["gid"] != want {
				t.Fatalf("gid changed: %s -> %s", raw, encoded)
			}
		}
		if count != 2 {
			t.Fatalf("managers = %d", count)
		}
	}
}

func TestManagerServicesRemainOnActiveInstance(t *testing.T) {
	for _, field := range []string{``, `,"services":null`, `,"services":{}`, `,"services":{"dashboard":"https://mgr:8443/","prometheus":"http://mgr:9283/"}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mgr": []byte(`{"available":true,"active_name":"a","active_addr":"v2:host:3300","standbys":[{"name":"b"}]` + field + `}`)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, row := range rows {
			if row.Kind != "mgr" {
				continue
			}
			count++
			mgr := row.Payload.(cephdomain.Manager)
			if !mgr.Active && mgr.Services != nil {
				t.Fatal("standby inherited services")
			}
			if mgr.Active {
				if field == `` || field == `,"services":null` {
					if mgr.Services != nil {
						t.Fatal("invented services")
					}
				} else if mgr.Services == nil {
					t.Fatal("lost service map")
				}
				if len(mgr.Services) > 0 && (mgr.Services["dashboard"] != "https://mgr:8443/" || mgr.Services["prometheus"] != "http://mgr:9283/") {
					t.Fatal("service URI changed")
				}
			}
		}
		if count != 2 {
			t.Fatalf("managers=%d", count)
		}
	}
}

func TestManagerAvailabilityRetainsUnknown(t *testing.T) {
	for _, field := range []string{``, `,"available":null`, `,"available":false`, `,"available":true`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mgr": []byte(`{"active_name":"a","standbys":[{"name":"b"}]` + field + `}`)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, row := range rows {
			if row.Kind != "mgr" {
				continue
			}
			count++
			encoded, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			var want any
			if field == `,"available":false` {
				want = false
			}
			if field == `,"available":true` {
				want = true
			}
			if got["available"] != want {
				t.Fatalf("field=%s got=%s", field, encoded)
			}
		}
		if count != 2 {
			t.Fatalf("manager count=%d", count)
		}
	}
}

func TestManagerServiceValuesRequireExplicitStrings(t *testing.T) {
	for _, raw := range []string{`{"dashboard":null}`, `{"dashboard":false}`, `{"dashboard":1}`, `{"dashboard":{}}`, `{"":"https://host/"}`, `{" dashboard":"https://host/"}`, `{"dashboard":""}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.mgr": []byte(`{"active_name":"a","standbys":[],"services":` + raw + `}`)}}}
		rows, err := p.Collect(context.Background(), ClusterAccess{}, "topology")
		if raw != `{"dashboard":""}` {
			if err == nil || len(rows) != 0 {
				t.Fatalf("accepted invalid service %s: %v", raw, err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "mgr" {
				continue
			}
			value, exists := row.Payload.(cephdomain.Manager).Services["dashboard"]
			if !exists || value != "" {
				t.Fatal("lost explicit empty URI")
			}
			found = true
		}
		if !found {
			t.Fatal("manager missing")
		}
	}
}
