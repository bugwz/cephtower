package clusterinspect

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestCephFSPoolUsageUsesLogicalStorage(t *testing.T) {
	service, runner, id := testInspection(t)
	runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
		if spec.ID == "cephfs.pools.map" {
			return executor.CommandResult{Stdout: []byte(`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[2,3,4]}}`)}, nil
		}
		return executor.CommandResult{Stdout: []byte(`{"pools":[{"id":8,"name":"unrelated","stats":{"stored":7,"max_avail":9}},{"id":2,"name":"cephfs.data","stats":{"stored":18446744073709551615,"max_avail":18446744073709551615,"bytes_used":99}},{"id":1,"name":"cephfs.meta","stats":{"stored":0,"max_avail":100,"bytes_used":12}},{"id":4,"name":"incomplete","stats":{"bytes_used":10}}]}`)}, nil
	}
	result, err := service.CephFSPools(context.Background(), id, "cephfs")
	if err != nil || len(result.Items) != 4 || result.ObservedAt.IsZero() {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
	meta, data, missing, incomplete := result.Items[0], result.Items[1], result.Items[2], result.Items[3]
	if meta.Type != "metadata" || meta.Name != "cephfs.meta" || *meta.Stored != "0" || *meta.Size != "100" || *meta.BytesUsed != "12" {
		t.Fatalf("incorrect metadata stats: %+v", meta)
	}
	if data.Type != "data" || *data.Stored != "18446744073709551615" || *data.Size != "36893488147419103230" || *data.BytesUsed != "99" {
		t.Fatalf("logical capacity lost precision or used physical bytes: %+v", data)
	}
	if missing.ID != "3" || missing.Error == "" || missing.Stored != nil || missing.Size != nil || incomplete.Error == "" || incomplete.Stored != nil {
		t.Fatalf("missing stats were fabricated: %+v %+v", missing, incomplete)
	}
	if len(runner.specs) != 2 || !reflect.DeepEqual(runner.specs[0].Args, []string{"fs", "get", "cephfs", "--format", "json"}) || !reflect.DeepEqual(runner.specs[1].Args, []string{"df", "detail", "--format", "json"}) || runner.specs[0].Mutating || runner.specs[1].Mutating {
		t.Fatalf("unexpected native chain: %+v", runner.specs)
	}
}

func TestCephFSPoolUsageValidatesMapAndDF(t *testing.T) {
	for _, tt := range []struct{ fs, df string }{
		{`null`, `{}`}, {`{}`, `{}`},
		{`{"mdsmap":{"fs_name":"other","metadata_pool":1,"data_pools":[]}}`, `{}`},
		{`{"mdsmap":{"fs_name":"cephfs","data_pools":[]}}`, `{}`},
		{`{"mdsmap":{"fs_name":"cephfs","metadata_pool":-1,"data_pools":[]}}`, `{}`},
		{`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[1]}}`, `{}`},
		{`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[2,-1]}}`, `{}`},
		{`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[]}}`, `null`},
		{`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[]}}`, `{"pools":[{}]}`},
		{`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[]}}`, `{"pools":[{"id":1},{"id":1}]}`},
	} {
		service, runner, id := testInspection(t)
		runner.run = func(spec executor.CommandSpec) (executor.CommandResult, error) {
			output := tt.df
			if spec.ID == "cephfs.pools.map" {
				output = tt.fs
			}
			return executor.CommandResult{Stdout: []byte(output)}, nil
		}
		if _, err := service.CephFSPools(context.Background(), id, "cephfs"); err == nil {
			t.Fatalf("accepted invalid data: %+v", tt)
		}
	}
	service, runner, id := testInspection(t)
	if _, err := service.CephFSPools(context.Background(), id, "--help"); err == nil || len(runner.specs) != 0 {
		t.Fatal("invalid filesystem executed")
	}
	runner.fail = true
	if _, err := service.CephFSPools(context.Background(), id, "cephfs"); err == nil {
		t.Fatal("command failure became empty usage")
	}
}

func TestUnsignedPoolStatPreservesMissingAndInvalid(t *testing.T) {
	for _, value := range []string{"", "-1", "1.2", "18446744073709551616"} {
		if unsignedPoolStat(json.Number(value)) != nil {
			t.Fatalf("invalid stats accepted: %s", value)
		}
	}
}
