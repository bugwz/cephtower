package ceph

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCephUserCollectionNeverPersistsKeys(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.ceph_user": []byte(`{"auth_dump":[{"entity":"client.backup","key":"sensitive-value","caps":{"mon":"allow r","osd":"profile rbd pool=data"}}]}`),
	}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "ceph_auth")
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v, error=%v", rows, err)
	}
	if rows[0].Kind != "ceph_user" || rows[0].NaturalKey != "client.backup" {
		t.Fatalf("row=%+v", rows[0])
	}
	value, _ := json.Marshal(rows)
	if strings.Contains(string(value), "sensitive-value") || strings.Contains(string(value), `"key"`) {
		t.Fatal("secret entered observation")
	}
	if !strings.Contains(string(value), "profile rbd pool=data") {
		t.Fatal("capabilities lost")
	}
}

func TestCephUserCollectionRejectsIncompleteData(t *testing.T) {
	for _, payload := range []string{`{}`, `{"auth_dump":null}`, `{"auth_dump":[{"caps":{}}]}`, `{"auth_dump":[{"entity":"client.a"}]}`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.ceph_user": []byte(payload)}}}
		if _, err := provider.Collect(context.Background(), ClusterAccess{}, "ceph_auth"); err == nil {
			t.Fatalf("accepted %s", payload)
		}
	}
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.ceph_user": []byte(`{"auth_dump":[]}`)}}}
	if rows, err := provider.Collect(context.Background(), ClusterAccess{}, "ceph_auth"); err != nil || len(rows) != 0 {
		t.Fatalf("valid empty collection rejected: %v", err)
	}
}
