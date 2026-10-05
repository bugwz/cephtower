package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type placementExecutor struct {
	calls     []executor.CommandSpec
	bodies    map[string]string
	fail      string
	codeStage string
}

func (e *placementExecutor) Run(_ context.Context, _ executor.ClusterAccess, c executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, c)
	stage := strings.TrimPrefix(c.ID, "rgw_zone.placement.")
	stage = strings.TrimPrefix(stage, "rgw_zone.placement_create.")
	stage = strings.TrimPrefix(stage, "rgw_zone.storage_class_create.")
	stage = strings.TrimPrefix(stage, "rgw_zonegroup.storage_class_create.")
	stage = strings.TrimPrefix(stage, "rgw_zonegroup.placement_create.")
	stage = strings.TrimPrefix(stage, "rgw_zonegroup.placement_default.")
	stage = strings.TrimPrefix(stage, "rgw_zonegroup.placement_tags.")
	stage = strings.TrimPrefix(stage, "rgw_zonegroup.storage_class_delete.")
	if stage == e.fail {
		return executor.CommandResult{}, errors.New("private key")
	}
	code := 0
	if stage == e.codeStage {
		code = 5
	}
	return executor.CommandResult{Stdout: []byte(e.bodies[stage]), ExitCode: code}, nil
}
func placementParams(realm string) map[string]any {
	return map[string]any{"zone_id": "z", "name": "zone", "realm_id": realm, "zonegroup_id": "g", "placement_id": "default", "storage_class": "COLD", "index_pool": "index:new", "data_pool": "data:new", "data_extra_pool": "", "compression": "zstd", "confirm_placement": true}
}
func placementFixture(realm string) *placementExecutor {
	zone := `{"id":"z","name":"zone","realm_id":"` + realm + `","system_key":{"secret_key":"private"},"placement_pools":[{"key":"default","val":{"index_pool":"index:old","data_extra_pool":"extra","index_type":0,"inline_data":true,"storage_classes":{"STANDARD":{"data_pool":"standard"},"COLD":{"data_pool":"data:old","compression_type":"none"}}}}]}`
	after := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(zone, "index:old", "index:new"), "data:old", "data:new"), `"data_extra_pool":"extra"`, `"data_extra_pool":""`), `"compression_type":"none"`, `"compression_type":"zstd"`)
	group := `{"id":"g","realm_id":"` + realm + `","zones":[{"id":"z"}],"placement_targets":[{"name":"default","storage_classes":["STANDARD","COLD"]}]}`
	period := `{"id":"new","realm_id":"r","epoch":2}`
	return &placementExecutor{bodies: map[string]string{"zone_before": zone, "zone_recheck": zone, "modify": after, "zone_after": after, "group_before": group, "group_recheck": group, "group_after": group, "realm_before": `{"id":"r","current_period":"old"}`, "period.pre_check": `{"id":"r","current_period":"old"}`, "period.commit": period, "period.realm_post_check": `{"id":"r","current_period":"new"}`, "period.period_post_check": period}}
}
func TestZonePlacementChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		runner := placementFixture(realm)
		s.executor = runner
		req := Request{ClusterID: cluster, Action: "rgw_zone.placement", Parameters: placementParams(realm)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if result.Details.(map[string]any)["data_migrated"] != false {
			t.Fatal("claimed migration")
		}
		count := 0
		for _, c := range runner.calls {
			if c.Binary != executor.BinaryRGWAdmin {
				t.Fatal("wrong binary")
			}
			if c.Mutating {
				count++
				if count == 1 {
					want := []string{"zone", "placement", "modify", "--zone-id", "z", "--zonegroup-id", "g", "--placement-id", "default", "--storage-class", "COLD", "--index-pool", "index:new", "--data-pool", "data:new", "--data-extra-pool", "", "--compression", "zstd", "--format", "json"}
					if !reflect.DeepEqual(c.Args, want) {
						t.Fatalf("wrong command %v", c.Args)
					}
				}
			}
		}
		wantCount := 1
		if realm != "" {
			wantCount = 2
		}
		if count != wantCount {
			t.Fatal("unexpected write count")
		}
		for _, call := range runner.calls {
			stage := strings.TrimPrefix(call.ID, "rgw_zone.placement.")
			for _, mode := range []string{"error", "exit"} {
				broken := placementFixture(realm)
				if mode == "error" {
					broken.fail = stage
				} else {
					broken.codeStage = stage
				}
				s.executor = broken
				_, err := s.Execute(context.Background(), req)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe error %s: %v", stage, err)
				}
				if broken.calls[len(broken.calls)-1].ID != call.ID {
					t.Fatal("continued after failed stage")
				}
			}
		}
	}
}
func TestZonePlacementPreflightAndPreservation(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tc := range []struct{ stage, from, to string }{
		{"zone_before", `"name":"zone"`, `"name":"renamed"`},
		{"group_before", `"zones":[{"id":"z"}]`, `"zones":[{"id":"other"}]`},
		{"group_before", `"storage_classes":["STANDARD","COLD"]`, `"storage_classes":["STANDARD"]`},
		{"group_before", `"name":"default"`, `"name":"default","tier_targets":[{"key":"COLD","val":{}}]`},
		{"zone_before", `"COLD":{`, `"OTHER":{`},
		{"zone_recheck", `"inline_data":true`, `"inline_data":false`},
		{"modify", `"inline_data":true`, `"inline_data":false`},
		{"zone_after", `"data_pool":"standard"`, `"data_pool":"unexpected"`},
		{"group_after", `"realm_id":"r"`, `"realm_id":"other"`},
	} {
		runner := placementFixture("r")
		runner.bodies[tc.stage] = strings.Replace(runner.bodies[tc.stage], tc.from, tc.to, 1)
		s.executor = runner
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zone.placement", Parameters: placementParams("r")})
		if err == nil {
			t.Fatalf("accepted drift %s", tc.stage)
		}
		if strings.Contains(tc.stage, "before") || strings.Contains(tc.stage, "recheck") {
			for _, call := range runner.calls {
				if call.Mutating {
					t.Fatal("wrote despite invalid preflight")
				}
			}
		}
	}
	for _, key := range []string{"zone_id", "realm_id", "index_pool", "data_pool", "data_extra_pool", "compression", "confirm_placement"} {
		p := placementParams("")
		delete(p, key)
		if _, err := buildZonePlacement(p); err == nil {
			t.Fatalf("missing %s accepted", key)
		}
	}
	for _, value := range []string{"a:b:c", `a\q`, "pool\n"} {
		p := placementParams("")
		p["data_pool"] = value
		if _, err := buildZonePlacement(p); err == nil {
			t.Fatal("noncanonical pool accepted")
		}
	}
	p := placementParams("")
	p["data_pool"] = `pool\:name:namespace`
	if _, err := buildZonePlacement(p); err != nil {
		t.Fatal(err)
	}
}
