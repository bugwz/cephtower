package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"fmt"
	"reflect"
	"testing"
)

type nfsInfoExecutor struct {
	data  string
	fail  bool
	calls []executor.CommandSpec
}

func (e *nfsInfoExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	if e.fail {
		return executor.CommandResult{}, fmt.Errorf("unavailable")
	}
	return executor.CommandResult{Stdout: []byte(e.data)}, nil
}
func TestNFSClusterInfo(t *testing.T) {
	for _, tc := range []struct {
		data            string
		fail, available bool
	}{
		{`{"a":{"virtual_ip":"2001:db8::1","port":2049,"monitor_port":9049,"ingress_mode":"haproxy-standard","backend":[{"hostname":"host","ip":"10.0.0.1","port":12049}]}}`, false, true},
		{`{"a":{"virtual_ip":null,"backend":[]}}`, false, true},
		{`{"a":{"backend":[{"hostname":"host","ip":"2001:db8::2","port":null}]}}`, false, true},
		{`{"a":{"backend":[{}]}}`, false, false},
		{`{"a":{"backend":[{"hostname":"host","ip":"bad","port":2049}]}}`, false, false},
		{`{"a":{"backend":[{"hostname":3,"ip":"10.0.0.1","port":2049}]}}`, false, false},
		{`{"a":{"backend":[{"hostname":"host","ip":"10.0.0.1","port":"2049"}]}}`, false, false},
		{`{"a":{"backend":[],"port":65536}}`, false, false},
		{`{"a":{"backend":[],"port":0}}`, false, false},
		{`{"a":{"backend":[],"monitor_port":1.5}}`, false, false},
		{`{"a":{"backend":[],"virtual_ip":{}}}`, false, false},
		{`{"a":{"backend":[],"ingress_mode":[]}}`, false, false},
		{`{"other":{"backend":[]}}`, false, false},
		{`{"a":{"backend":[null]}}`, false, false},
		{`{"a":null}`, false, false},
		{`invalid`, false, false},
		{`{}`, true, false},
	} {
		runner := &nfsInfoExecutor{data: tc.data, fail: tc.fail}
		provider := NativeProvider{Executor: runner}
		result := provider.collectNFSClusterInfo(context.Background(), ClusterAccess{}, "a")
		if result["name"] != "a" || result["info_available"] != tc.available {
			t.Fatalf("result=%v", result)
		}
		if len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0].Args, []string{"nfs", "cluster", "info", "a", "--format", "json"}) || runner.calls[0].Binary != executor.BinaryCeph {
			t.Fatal("incorrect native command")
		}
		if tc.available && result["backend"] == nil {
			t.Fatal("missing backends")
		}
		if !tc.available && result["backend"] != nil {
			t.Fatal("invalid data exposed")
		}
	}
}

func TestNFSClusterInfoFiltersUnexpectedFields(t *testing.T) {
	runner := &nfsInfoExecutor{data: `{"a":{"backend":[{"hostname":"host","ip":"10.0.0.1","port":2049,"secret":"hidden"}],"secret":"hidden"}}`}
	provider := NativeProvider{Executor: runner}
	result := provider.collectNFSClusterInfo(context.Background(), ClusterAccess{}, "a")
	if result["info_available"] != true || result["secret"] != nil {
		t.Fatal(result)
	}
	entry := result["backend"].([]map[string]any)[0]
	if entry["secret"] != nil || entry["hostname"] != "host" || entry["ip"] != "10.0.0.1" {
		t.Fatal(entry)
	}
}
