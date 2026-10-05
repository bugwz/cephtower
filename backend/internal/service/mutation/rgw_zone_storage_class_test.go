package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func storageClassFixture(realm string) *placementExecutor {
	runner := placementFixture(realm)
	for _, stage := range []string{"zone_before", "zone_recheck"} {
		runner.bodies[stage] = strings.Replace(runner.bodies[stage], `,"COLD":{"data_pool":"data:old","compression_type":"none"}`, "", 1)
	}
	for _, stage := range []string{"modify", "zone_after"} {
		runner.bodies[stage] = strings.ReplaceAll(strings.ReplaceAll(runner.bodies[stage], "index:new", "index:old"), `"data_extra_pool":""`, `"data_extra_pool":"extra"`)
	}
	return runner
}
func storageClassParams(realm string) map[string]any {
	p := placementParams(realm)
	delete(p, "index_pool")
	delete(p, "data_extra_pool")
	return p
}
func TestZoneStorageClassCreation(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		runner := storageClassFixture(realm)
		s.executor = runner
		req := Request{ClusterID: cluster, Action: "rgw_zone.storage_class_create", Parameters: storageClassParams(realm)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		details := result.Details.(map[string]any)
		if details["storage_class_created"] != true || details["period_published"] != false {
			t.Fatal("wrong creation result")
		}
		writes := 0
		for _, c := range runner.calls {
			if strings.Contains(c.ID, "period") || strings.Contains(c.ID, "realm_before") {
				t.Fatal("unexpected Realm publication")
			}
			if c.Mutating {
				writes++
				want := []string{"zone", "placement", "add", "--zone-id", "z", "--zonegroup-id", "g", "--placement-id", "default", "--storage-class", "COLD", "--data-pool", "data:new", "--compression", "zstd", "--format", "json"}
				if !reflect.DeepEqual(c.Args, want) {
					t.Fatalf("unexpected command %v", c.Args)
				}
			}
		}
		if writes != 1 {
			t.Fatal("unexpected writes")
		}
		for _, call := range runner.calls {
			stage := strings.TrimPrefix(call.ID, "rgw_zone.storage_class_create.")
			for _, mode := range []string{"exit", "error"} {
				broken := storageClassFixture(realm)
				if mode == "exit" {
					broken.codeStage = stage
				} else {
					broken.fail = stage
				}
				s.executor = broken
				_, err := s.Execute(context.Background(), req)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe failure %v", err)
				}
				if broken.calls[len(broken.calls)-1].ID != call.ID {
					t.Fatal("continued after failure")
				}
			}
		}
	}
}
func TestZoneStorageClassRejectsOverwriteAndDrift(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, scenario := range []string{"existing", "concurrent_creation", "wrong_group", "undeclared", "tier", "index_changed", "extra_changed", "other_class_changed"} {
		runner := storageClassFixture("r")
		switch scenario {
		case "existing":
			runner.bodies["zone_before"] = placementFixture("r").bodies["zone_before"]
		case "concurrent_creation":
			runner.bodies["zone_recheck"] = placementFixture("r").bodies["zone_before"]
		case "wrong_group":
			runner.bodies["group_before"] = `{}`
		case "undeclared":
			runner.bodies["group_before"] = strings.Replace(runner.bodies["group_before"], `,"COLD"`, "", 1)
		case "tier":
			runner.bodies["group_before"] = strings.Replace(runner.bodies["group_before"], `"name":"default"`, `"name":"default","tier_targets":[{"key":"COLD","val":{}}]`, 1)
		case "index_changed":
			runner.bodies["zone_after"] = strings.Replace(runner.bodies["zone_after"], "index:old", "wrong", 1)
		case "extra_changed":
			runner.bodies["zone_after"] = strings.Replace(runner.bodies["zone_after"], `"data_extra_pool":"extra"`, `"data_extra_pool":"wrong"`, 1)
		case "other_class_changed":
			runner.bodies["zone_after"] = strings.Replace(runner.bodies["zone_after"], `"data_pool":"standard"`, `"data_pool":"wrong"`, 1)
		}
		s.executor = runner
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zone.storage_class_create", Parameters: storageClassParams("r")})
		if err == nil {
			t.Fatalf("accepted %s", scenario)
		}
		if !strings.HasSuffix(scenario, "changed") {
			for _, call := range runner.calls {
				if call.Mutating {
					t.Fatalf("wrote for %s", scenario)
				}
			}
		}
	}
}
