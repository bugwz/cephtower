package clusterinspect

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

const rankMapFixture = `{"filesystems":[{"mdsmap":{"fs_name":"other","in":[],"up":{},"info":{}}},{"mdsmap":{"fs_name":"cephfs","in":[1,0],"up":{"mds_0":9007199254740993},"info":{"gid_9007199254740993":{"name":"active","gid":9007199254740993,"rank":0,"state":"up:active","laggy_since":"2026-10-01T12:00:00"},"gid_2":{"name":"replay","gid":2,"rank":0,"state":"up:standby-replay"}}}}],"standbys":[{"name":"spare","gid":3,"rank":-1,"state":"up:standby"}]}`

func TestCephFSMDSRankTopology(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID == "cephfs.mds.map" {
			return executor.CommandResult{Stdout: []byte(rankMapFixture)}, nil
		}
		return executor.CommandResult{Stdout: []byte(`[{"name":"active","ceph_version":"ceph version fixture"},{"name":"spare","ceph_version":"ceph version spare"},{"name":"unrelated","ceph_version":"ignored"}]`)}, nil
	}
	result, err := service.CephFSMDS(context.Background(), id, "cephfs")
	if err != nil || len(result.Ranks) != 3 || len(result.Standbys) != 1 || result.ObservedAt.IsZero() {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	active, failed, replay := result.Ranks[0], result.Ranks[1], result.Ranks[2]
	if active.Rank != "0" || !active.Laggy || active.GID != "9007199254740993" || active.State != "active" || active.Version != "ceph version fixture" {
		t.Fatalf("active = %+v", active)
	}
	if failed.Rank != "1" || failed.State != "failed" || failed.Name != "" || failed.GID != "" {
		t.Fatalf("failed = %+v", failed)
	}
	if replay.Rank != "0-s" || replay.Name != "replay" || replay.State != "standby-replay" || result.Standbys[0].Name != "spare" || result.Standbys[0].Version != "ceph version spare" {
		t.Fatalf("replay/standby = %+v", result)
	}
	if len(runner.specs) != 2 || !reflect.DeepEqual(runner.specs[0].Args, []string{"fs", "dump", "--format", "json"}) || !reflect.DeepEqual(runner.specs[1].Args, []string{"mds", "metadata", "--format", "json"}) || runner.specs[0].Mutating || runner.specs[1].Mutating {
		t.Fatalf("commands = %+v", runner.specs)
	}
}

func TestCephFSMDSValidatesRankRelationships(t *testing.T) {
	for _, value := range []string{
		`null`, `{}`, `{"filesystems":[],"standbys":[]}`,
		strings.Replace(rankMapFixture, `"in":[1,0]`, `"in":[0,0]`, 1),
		strings.Replace(rankMapFixture, `"mds_0":9007199254740993`, `"mds_0":404`, 1),
		strings.Replace(rankMapFixture, `"mds_0":9007199254740993`, `"mds_4":9007199254740993`, 1),
		strings.Replace(rankMapFixture, `"rank":0,"state":"up:active"`, `"rank":1,"state":"up:active"`, 1),
		strings.Replace(rankMapFixture, `"gid_2"`, `"gid_wrong"`, 1),
		strings.Replace(rankMapFixture, `"name":"replay"`, `"name":"active"`, 1),
		strings.Replace(rankMapFixture, `"rank":0,"state":"up:standby-replay"`, `"rank":4,"state":"up:standby-replay"`, 1),
		strings.Replace(rankMapFixture, `"rank":-1`, `"rank":0`, 1),
	} {
		service, runner, id := testInspection(t)
		runner.output = value
		if _, err := service.CephFSMDS(context.Background(), id, "cephfs"); err == nil {
			t.Fatalf("accepted invalid FS map: %s", value)
		}
	}
}

func TestCephFSMDSMetadataFailureDoesNotHideRanks(t *testing.T) {
	for _, value := range []string{`null`, `[]`, `[{"name":"active"},{"name":"active"}]`, `[{"ceph_version":"unknown"}]`, "failure"} {
		service, runner, id := testInspection(t)
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			if spec.ID == "cephfs.mds.map" {
				return executor.CommandResult{Stdout: []byte(rankMapFixture)}, nil
			}
			if value == "failure" {
				return executor.CommandResult{}, errors.New("token=fixture")
			}
			return executor.CommandResult{Stdout: []byte(value)}, nil
		}
		result, err := service.CephFSMDS(context.Background(), id, "cephfs")
		if err != nil || len(result.Ranks) != 3 || (value != `[]` && result.MetadataError == "") || strings.Contains(result.MetadataError, "fixture") {
			t.Fatalf("metadata failure lost topology: %+v %v", result, err)
		}
	}
}
