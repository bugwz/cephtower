package ceph

import (
	"context"
	"reflect"
	"testing"
	"time"
)

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
