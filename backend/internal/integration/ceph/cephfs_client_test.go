package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestCollectCephFSClientsUsesFilesystemRankAndDecoratesMetadata(t *testing.T) {
	raw := `{"sessions":[
		{"id":123,"state":"open","num_caps":7,"client_metadata":{"ceph_version":"ceph version 20.2.2","hostname":"node-a","root":"/projects"}},
		{"id":456,"state":"stale","client_metadata":{"kernel_version":"6.12.0","hostname":"node-b","root":"/"}}
	]}`
	var calls []executor.CommandSpec
	provider := NativeProvider{Executor: recordingExecutor{
		base:  malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.cephfs_client": []byte(raw)}},
		calls: &calls,
	}}
	now := time.Unix(123, 0).UTC()
	rows := provider.collectCephFSClients(context.Background(), ClusterAccess{}, "cephfs", now)
	if len(calls) != 1 || calls[0].Binary != executor.BinaryCeph || calls[0].Mutating || !reflect.DeepEqual(calls[0].Args, []string{"tell", "mds.cephfs:0", "session", "ls", "--format", "json"}) {
		t.Fatalf("command = %+v", calls)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %#v", rows)
	}
	first := rows[0]
	if first.NaturalKey != "cephfs/123" || first.ParentKind != "filesystem" || first.ParentKey != "cephfs" || first.Status != "open" || !first.ObservedAt.Equal(now) {
		t.Fatalf("first row = %+v", first)
	}
	payload := first.Payload.(map[string]any)
	if payload["client_id"] != "123" || payload["type"] != "userspace" || payload["version"] != "ceph version 20.2.2" || payload["hostname"] != "node-a" || payload["root"] != "/projects" || payload["num_caps"] != json.Number("7") {
		t.Fatalf("first payload = %#v", payload)
	}
	second := rows[1].Payload.(map[string]any)
	if second["type"] != "kernel" || second["version"] != "6.12.0" {
		t.Fatalf("second payload = %#v", second)
	}
}

func TestCephFSSessionListSupportsArrayAndRejectsInvalidShapes(t *testing.T) {
	items := []any{map[string]any{"id": json.Number("1")}}
	for _, payload := range []any{items, map[string]any{"sessions": items}, map[string]any{"sessions": []any{}}} {
		if _, ok := cephFSSessionList(payload); !ok {
			t.Fatalf("valid payload rejected: %#v", payload)
		}
	}
	for _, payload := range []any{nil, map[string]any{}, map[string]any{"sessions": nil}, []any{"invalid"}} {
		if _, ok := cephFSSessionList(payload); ok {
			t.Fatalf("invalid payload accepted: %#v", payload)
		}
	}
}
