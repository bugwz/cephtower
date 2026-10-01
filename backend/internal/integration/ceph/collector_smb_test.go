package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSMBAuthInventory(t *testing.T) {
	for _, spec := range []struct{ kind, resourceType, idField string }{
		{"smb_join_auth", "ceph.smb.join.auth", "auth_id"},
		{"smb_usersgroups", "ceph.smb.usersgroups", "users_groups_id"},
	} {
		t.Run(spec.kind, func(t *testing.T) {
			item := map[string]any{"resource_type": spec.resourceType, spec.idField: "auth-a", "intent": "present", "linked_to_cluster": "cluster-a", "auth": map[string]any{"password": "secret"}, "values": map[string]any{"users": []string{"secret"}}, "unexpected": "secret"}
			data, _ := json.Marshal(map[string]any{"resources": []any{item}})
			runner := &nfsInfoExecutor{data: string(data)}
			provider := NativeProvider{Executor: runner}
			rows := provider.collectSMBAuthResources(context.Background(), ClusterAccess{}, time.Time{})
			if len(rows) != 1 || rows[0].Kind != spec.kind || rows[0].NaturalKey != "auth-a" {
				t.Fatal(rows)
			}
			encoded, _ := json.Marshal(rows[0].Payload)
			if strings.Contains(string(encoded), "secret") || len(rows[0].Payload.(map[string]any)) != 4 {
				t.Fatal(string(encoded))
			}
			for _, call := range runner.calls {
				resourceType := "ceph.smb.join.auth"
				if call.ID == "collect.smb_usersgroups" {
					resourceType = "ceph.smb.usersgroups"
				}
				if call.Mutating || !reflect.DeepEqual(call.Args, []string{"smb", "show", resourceType, "--results=full", "--password-filter=hidden", "--format", "json"}) {
					t.Fatal(call)
				}
			}
		})
	}
}

func TestSMBAuthInventoryUnavailable(t *testing.T) {
	for _, data := range []string{`null`, `{}`, `{"resources":null}`, `invalid`, `{"resources":[null]}`, `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"bad/id"}]}`, `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"a"},{"resource_type":"ceph.smb.join.auth","auth_id":"a"}]}`, `{"resources":[{"resource_type":"ceph.smb.join.auth","auth_id":"a","linked_to_cluster":{}}]}`, `{"resources":[]}`} {
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		provider := NativeProvider{Executor: &nfsInfoExecutor{data: data}}
		if rows := provider.collectSMBAuthResources(ctx, ClusterAccess{}, time.Time{}); len(rows) != 0 {
			t.Fatal(rows)
		}
		want := 2
		if data == `{"resources":[]}` {
			want = 0
		}
		if len(trace.unavailable) != want {
			t.Fatalf("%s: %v", data, trace.unavailable)
		}
	}
}

func TestSMBClusterInfo(t *testing.T) {
	for _, tc := range []struct {
		data      string
		available bool
	}{
		{`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"active-directory","domain_settings":{"realm":"EXAMPLE.COM","join_sources":[{"ref":"auth-a"}]},"placement":{"count":2},"custom_dns":["192.0.2.1"],"unexpected":"hidden"}`, true},
		{`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":"user"}`, true},
		{`{"resource_type":"ceph.smb.cluster","cluster_id":"other","auth_mode":"user"}`, false},
		{`{"resource_type":"ceph.smb.share","cluster_id":"a","auth_mode":"user"}`, false},
		{`{"resource_type":"ceph.smb.cluster","cluster_id":"a","auth_mode":null}`, false},
		{`{"resources":[]}`, false}, {`null`, false}, {`invalid`, false},
	} {
		runner := &nfsInfoExecutor{data: tc.data}
		provider := NativeProvider{Executor: runner}
		got := provider.collectSMBClusterInfo(context.Background(), ClusterAccess{}, "a")
		if got["name"] != "a" || got["info_available"] != tc.available || got["unexpected"] != nil {
			t.Fatal(got)
		}
		if !reflect.DeepEqual(runner.calls[0].Args, []string{"smb", "show", "ceph.smb.cluster.a", "--format", "json"}) || runner.calls[0].Mutating {
			t.Fatal(runner.calls)
		}
		if tc.available && got["auth_mode"] == "active-directory" && got["domain_settings"].(map[string]any)["realm"] != "EXAMPLE.COM" {
			t.Fatal("domain configuration lost")
		}
	}
}

func TestSMBShareResources(t *testing.T) {
	runner := &nfsInfoExecutor{data: `{"resources":[{"resource_type":"ceph.smb.share","cluster_id":"a","share_id":"docs","name":"Documents","readonly":false,"browseable":true,"cephfs":{"volume":"fs","path":"/docs"}}]}`}
	provider := NativeProvider{Executor: runner}
	rows := provider.collectSMBShares(context.Background(), ClusterAccess{}, "a", time.Time{})
	if len(rows) != 1 || rows[0].NaturalKey != opaquePair("a", "docs") || rows[0].Payload.(map[string]any)["name"] != "Documents" {
		t.Fatal(rows)
	}
	if !reflect.DeepEqual(runner.calls[0].Args, []string{"smb", "show", "ceph.smb.share.a", "--results=full", "--format", "json"}) {
		t.Fatal(runner.calls)
	}
	for _, data := range []string{`["docs"]`, `{"resources":null}`, `{"resources":[{"resource_type":"ceph.smb.share","cluster_id":"other","share_id":"docs"}]}`, `{"resources":[null]}`} {
		runner.data = data
		if rows := provider.collectSMBShares(context.Background(), ClusterAccess{}, "a", time.Time{}); len(rows) != 0 {
			t.Fatal(rows)
		}
	}
}
