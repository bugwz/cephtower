package clusterinspect

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestCephFSCountersNativeScopeAndPrecision(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID == "cephfs.performance.map" {
			return executor.CommandResult{Stdout: []byte(`{"mdsmap":{"fs_name":"cephfs","info":{"gid_b":{"name":"b","gid":9007199254740993,"rank":1,"state":"up:standby-replay"},"gid_a":{"name":"a","gid":1,"rank":0,"state":"up:active"}}}}`)}, nil
		}
		if spec.Args[1] == "mds.b" {
			return executor.CommandResult{}, errors.New("token=secret-fixture")
		}
		return executor.CommandResult{Stdout: []byte(`{"mds_server":{"handle_client_request":9007199254740993},"mds_mem":{"ino":0}}`)}, nil
	}
	result, err := service.CephFSCounters(context.Background(), id, "cephfs")
	if err != nil || len(result.Items) != 2 {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	a, b := result.Items[0], result.Items[1]
	if a.Name != "a" || a.State != "up:active" || len(a.Counters) != 11 || *a.Counters[0].Value != "9007199254740993" || *a.Counters[10].Value != "0" || a.Counters[1].Value != nil || a.ObservedAt.IsZero() {
		t.Fatalf("invalid sample: %+v", a)
	}
	if b.GID != "9007199254740993" || b.Error != "token=[REDACTED]" || len(b.Counters) != 0 {
		t.Fatalf("failed daemon was misrepresented: %+v", b)
	}
	if len(runner.specs) != 3 || !reflect.DeepEqual(runner.specs[0].Args, []string{"fs", "get", "cephfs", "--format", "json"}) || !reflect.DeepEqual(runner.specs[1].Args, []string{"tell", "mds.a", "perf", "dump", "--format", "json"}) {
		t.Fatalf("invalid scope: %+v", runner.specs)
	}
	for _, spec := range runner.specs {
		if spec.Mutating {
			t.Fatal("performance read was marked mutating")
		}
	}
}

func TestCephFSCountersRejectsInvalidMaps(t *testing.T) {
	service, runner, id := testInspection(t)
	for _, value := range []string{
		`{}`, `null`, `{"mdsmap":{"fs_name":"other","info":{}}}`,
		`{"mdsmap":{"fs_name":"cephfs"}}`,
		`{"mdsmap":{"fs_name":"cephfs","info":{"a":{"name":"a,*","gid":1}}}}`,
		`{"mdsmap":{"fs_name":"cephfs","info":{"a":{"name":"a","gid":-1}}}}`,
		`{"mdsmap":{"fs_name":"cephfs","info":{"a":{"name":"a","gid":1},"b":{"name":"a","gid":2}}}}`,
	} {
		runner.output = value
		if _, err := service.CephFSCounters(context.Background(), id, "cephfs"); err == nil {
			t.Fatalf("accepted invalid map: %s", value)
		}
	}
	runner.output = `{"mdsmap":{"fs_name":"cephfs","info":{}}}`
	if result, err := service.CephFSCounters(context.Background(), id, "cephfs"); err != nil || result.Items == nil || len(result.Items) != 0 {
		t.Fatalf("empty valid map: %+v, %v", result, err)
	}
	runner.specs = nil
	if _, err := service.CephFSCounters(context.Background(), id, "--help"); err == nil || len(runner.specs) != 0 {
		t.Fatal("invalid filesystem executed a command")
	}
}

func TestCephFSCountersMarksMalformedDaemonData(t *testing.T) {
	for _, value := range []string{`null`, `{}`, `[]`, `{"mds_mem":{"ino":"3"}}`, `{"mds_mem":{"ino":-1}}`, `{"mds_mem":{"ino":1.2}}`, `{"mds_mem":{"ino":18446744073709551616}}`, `{"mds_mem":{"ino":{"avgcount":1}}}`} {
		service, runner, id := testInspection(t)
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if spec.ID == "cephfs.performance.map" {
				return executor.CommandResult{Stdout: []byte(`{"mdsmap":{"fs_name":"cephfs","info":{"a":{"name":"a","gid":1,"rank":0,"state":"up:active"}}}}`)}, nil
			}
			return executor.CommandResult{Stdout: []byte(value)}, nil
		}
		result, err := service.CephFSCounters(context.Background(), id, "cephfs")
		if err != nil || len(result.Items) != 1 || result.Items[0].Error == "" || len(result.Items[0].Counters) != 0 {
			t.Fatalf("malformed data became valid counters: %s %+v %v", value, result, err)
		}
	}
}
