package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
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

func TestCephUserCollectionDiscoversCephFSAuthorizations(t *testing.T) {
	payload := `{"auth_dump":[
		{"entity":"client.app","key":"sensitive-value","caps":{"mds":"allow rwps fsname=cephfs path='shared projects' root_squash uid=1000 gids=1000,1001 network 10.0.0.0/8, allow r fsname=archive path=/reports","mon":"allow r"}},
		{"entity":"client.global","caps":{"mds":"allow *"}},
		{"entity":"osd.0","caps":{"mds":"allow rw fsname=ignored"}},
		{"entity":"client.invalid","caps":{"mds":"allow w fsname=ignored"}}
	]}`
	var calls []executor.CommandSpec
	provider := NativeProvider{Executor: recordingExecutor{
		base:  malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.ceph_user": []byte(payload)}},
		calls: &calls,
	}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "ceph_auth")
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0].Binary != executor.BinaryCeph || calls[0].Mutating || !reflect.DeepEqual(calls[0].Args, []string{"auth", "ls", "--format", "json"}) {
		t.Fatalf("command = %+v", calls)
	}
	if len(rows) != 7 {
		t.Fatalf("rows = %d, want 7", len(rows))
	}
	var authorizations []Observation
	for _, row := range rows {
		if row.Kind == "cephfs_authorization" {
			authorizations = append(authorizations, row)
		}
	}
	if len(authorizations) != 3 {
		t.Fatalf("authorizations = %#v", authorizations)
	}
	first := authorizations[0]
	details, ok := first.Payload.(map[string]any)
	if !ok {
		t.Fatalf("payload type = %T", first.Payload)
	}
	if first.ParentKind != "filesystem" || first.ParentKey != "cephfs" || first.Name != "client.app /shared projects" {
		t.Fatalf("authorization = %+v", first)
	}
	wantDetails := map[string]any{
		"fs": "cephfs", "client": "client.app", "path": "/shared projects", "access": "rw", "permissions": "rwps",
		"quota": true, "snapshot": true, "full": false, "root_squash": true,
		"uid": "1000", "gids": "1000,1001", "network": "10.0.0.0/8",
		"mds_cap": "allow rwps fsname=cephfs path='shared projects' root_squash uid=1000 gids=1000,1001 network 10.0.0.0/8",
	}
	if !reflect.DeepEqual(details, wantDetails) {
		t.Fatalf("payload = %#v, want %#v", details, wantDetails)
	}
	global := authorizations[2]
	globalDetails := global.Payload.(map[string]any)
	if global.ParentKey != "*" || globalDetails["path"] != "/" || globalDetails["access"] != "*" || globalDetails["quota"] != true || globalDetails["snapshot"] != true || globalDetails["full"] != true {
		t.Fatalf("global authorization = %+v", global)
	}
	encoded, _ := json.Marshal(rows)
	if strings.Contains(string(encoded), "sensitive-value") || strings.Contains(string(encoded), `"key"`) {
		t.Fatal("secret entered authorization observation")
	}
}

func TestParseMDSAuthorizationGrantsSupportsCephSeparatorsAndQuotedPaths(t *testing.T) {
	grants := parseMDSAuthorizationGrants(`allow r fsname=a path="team, blue"; allow rwf fsname=b path=dir uid=42 gids=1,2`)
	if len(grants) != 2 {
		t.Fatalf("grants = %#v", grants)
	}
	if grants[0].Path != "/team, blue" || grants[1].Filesystem != "b" || grants[1].Path != "/dir" || grants[1].UID != "42" || grants[1].GIDs != "1,2" {
		t.Fatalf("grants = %#v", grants)
	}
	if got := parseMDSAuthorizationGrants(`allow rw fsname=a path="unterminated`); len(got) != 0 {
		t.Fatalf("malformed cap accepted: %#v", got)
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
