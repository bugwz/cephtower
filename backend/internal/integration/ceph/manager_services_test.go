package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"testing"
)

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
