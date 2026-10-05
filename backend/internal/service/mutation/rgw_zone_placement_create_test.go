package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"strings"
	"testing"
)

func placementCreateFixture(realm string) *placementExecutor {
	r := placementFixture(realm)
	entry := `{"key":"new","val":{"index_pool":"index:new","data_extra_pool":"","index_type":0,"inline_data":true,"storage_classes":{"STANDARD":{"data_pool":"data:new","compression_type":"zstd"}}}}`
	after := strings.TrimSuffix(r.bodies["zone_before"], `]}`) + `,` + entry + `]}`
	r.bodies["modify"], r.bodies["zone_after"] = after, after
	for _, stage := range []string{"group_before", "group_recheck", "group_after"} {
		r.bodies[stage] = strings.Replace(r.bodies[stage], `"name":"default"`, `"name":"new"`, 1)
	}
	return r
}
func placementCreateParams(realm string) map[string]any {
	p := placementParams(realm)
	delete(p, "storage_class")
	p["placement_id"] = "new"
	return p
}
func TestZonePlacementCreationChain(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		r := placementCreateFixture(realm)
		s.executor = r
		req := Request{ClusterID: cluster, Action: "rgw_zone.placement_create", Parameters: placementCreateParams(realm)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		details := result.Details.(map[string]any)
		if details["placement_created"] != true || details["period_published"] != (realm != "") {
			t.Fatal(details)
		}
		if _, ok := req.Parameters["storage_class"]; ok {
			t.Fatal("mutated request")
		}
		for _, call := range r.calls {
			if strings.HasSuffix(call.ID, ".modify") && strings.Join(call.Args, " ") != "zone placement add --zone-id z --zonegroup-id g --placement-id new --storage-class STANDARD --index-pool index:new --data-pool data:new --data-extra-pool  --compression zstd --format json" {
				t.Fatal(call.Args)
			}
			for _, mode := range []string{"exit", "error"} {
				broken := placementCreateFixture(realm)
				stage := strings.TrimPrefix(call.ID, "rgw_zone.placement_create.")
				if mode == "exit" {
					broken.codeStage = stage
				} else {
					broken.fail = stage
				}
				s.executor = broken
				_, err := s.Execute(context.Background(), req)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe error %v", err)
				}
				if broken.calls[len(broken.calls)-1].ID != call.ID {
					t.Fatal("continued after failure")
				}
			}
		}
	}
}
func TestZonePlacementCreationRejectsOverwrite(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, stage := range []string{"zone_before", "zone_recheck", "group_before", "zone_after"} {
		r := placementCreateFixture("r")
		if stage == "group_before" {
			r.bodies[stage] = `{}`
		} else if stage == "zone_after" {
			r.bodies[stage] = strings.Replace(r.bodies[stage], "index:old", "wrong", 1)
		} else {
			r.bodies[stage] = r.bodies["modify"]
		}
		s.executor = r
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zone.placement_create", Parameters: placementCreateParams("r")})
		if err == nil {
			t.Fatal(stage)
		}
		if stage != "zone_after" {
			for _, call := range r.calls {
				if call.Mutating {
					t.Fatal("wrote before validation")
				}
			}
		}
	}
	expected, ok := zonePlacementCreateExpected(periodDocument([]byte(`{"placement_pools":[]}`)), placementCreateParams(""))
	if !ok || len(expected["placement_pools"].([]any)) != 1 {
		t.Fatal("cannot initialize empty placements")
	}
}
