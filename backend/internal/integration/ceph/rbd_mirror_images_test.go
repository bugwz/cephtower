package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMirrorImageSiteStatusPreserved(t *testing.T) {
	const raw = `{"summary":{"health":"WARNING"},"daemons":[],"images":[{"name":"image","global_id":"global","state":"up+replaying","description":"replaying","last_update":"2026-01-01 12:00:01","daemon_service":{"service_id":"svc","instance_id":"9007199254740993","daemon_id":"client.mirror","hostname":"host"},"peer_sites":[{"site_name":"","mirror_uuid":"remote","state":"down+unknown","description":"<native text>","last_update":"2026-01-01 12:00:00"}]},{"name":"partial","global_id":"partial-id"}]}`
	var expected map[string]any
	if err := json.Unmarshal([]byte(raw), &expected); err != nil {
		t.Fatal(err)
	}
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rbd_mirroring":        []byte(`{"mode":"image","peers":[]}`),
		"collect.rbd_mirroring_status": []byte(raw),
	}}}
	rows := provider.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool-a"}}, fsDumpWire{}, time.Now())
	for _, row := range rows {
		if row.Kind == "rbd_mirroring" {
			payload := row.Payload.(map[string]any)
			if row.NaturalKey != "pool-a" || !reflect.DeepEqual(payload["images"], expected["images"]) {
				t.Fatalf("changed mirror image status: %#v", payload)
			}
			return
		}
	}
	t.Fatal("missing pool mirror inventory")
}
