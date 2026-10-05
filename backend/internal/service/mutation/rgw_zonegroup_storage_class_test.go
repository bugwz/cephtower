package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func groupClassParams(def string) map[string]any {
	return map[string]any{"zonegroup_id": "g", "name": "group", "realm_id": "r", "placement_id": "p", "storage_class": "COLD", "expected_default_placement": def, "confirm_create": true}
}
func groupClassFixture(def string) *placementExecutor {
	group := `{"id":"g","name":"group","realm_id":"r","default_placement":"` + def + `","zones":[{"id":"z"}],"placement_targets":[{"name":"p","tags":["restricted"],"storage_classes":["STANDARD"]},{"name":"other","tags":[],"storage_classes":["STANDARD"]}]}`
	after := strings.Replace(group, `"storage_classes":["STANDARD"]`, `"storage_classes":["COLD","STANDARD"]`, 1)
	if def == "" {
		after = strings.Replace(after, `"default_placement":""`, `"default_placement":"p"`, 1)
	}
	return &placementExecutor{bodies: map[string]string{"before": group, "recheck": group, "add": `[]`, "after": after}}
}
func TestZonegroupStorageClassDeclaration(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, def := range []string{"", "other"} {
		runner := groupClassFixture(def)
		s.executor = runner
		req := Request{ClusterID: cluster, Action: "rgw_zonegroup.storage_class_create", Parameters: groupClassParams(def)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		d := result.Details.(map[string]any)
		if d["default_placement_initialized"] != (def == "") || d["period_published"] != false {
			t.Fatal("wrong declaration result")
		}
		if len(runner.calls) != 4 {
			t.Fatal("unexpected call count")
		}
		wanted := []string{"zonegroup", "placement", "add", "--zonegroup-id", "g", "--placement-id", "p", "--storage-class", "COLD", "--format", "json"}
		for i, call := range runner.calls {
			if call.Mutating != (i == 2) {
				t.Fatal("unexpected write")
			}
			if i == 2 && !reflect.DeepEqual(call.Args, wanted) {
				t.Fatal("incorrect add scope")
			}
		}
		for _, stage := range []string{"before", "recheck", "add", "after"} {
			for _, mode := range []string{"error", "exit"} {
				broken := groupClassFixture(def)
				if mode == "error" {
					broken.fail = stage
				} else {
					broken.codeStage = stage
				}
				s.executor = broken
				_, err := s.Execute(context.Background(), req)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatalf("unsafe error %v", err)
				}
				if broken.calls[len(broken.calls)-1].ID != req.Action+"."+stage {
					t.Fatal("continued after failure")
				}
			}
		}
	}
}
func TestZonegroupStorageClassRejectsConflictAndCollateralChange(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tc := range []struct{ stage, from, to string }{
		{"before", `"default_placement":"other"`, `"default_placement":"changed"`},
		{"before", `"storage_classes":["STANDARD"]`, `"storage_classes":["COLD","STANDARD"]`},
		{"before", `"name":"p"`, `"name":"missing"`},
		{"before", `"name":"p"`, `"name":"p","tier_targets":[{"key":"COLD","val":{}}]`},
		{"recheck", `"restricted"`, `"changed"`},
		{"after", `"restricted"`, `"changed"`},
		{"after", `"default_placement":"other"`, `"default_placement":"p"`},
		{"after", `"id":"z"`, `"id":"unexpected"`},
	} {
		runner := groupClassFixture("other")
		runner.bodies[tc.stage] = strings.Replace(runner.bodies[tc.stage], tc.from, tc.to, 1)
		s.executor = runner
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.storage_class_create", Parameters: groupClassParams("other")})
		if err == nil {
			t.Fatalf("accepted %s", tc.stage)
		}
		if tc.stage != "after" {
			for _, call := range runner.calls {
				if call.Mutating {
					t.Fatal("wrote despite conflict")
				}
			}
		}
	}
	p := groupClassParams("")
	p["placement_id"] = "p/OTHER"
	if _, err := buildZonegroupStorageClass(p); err == nil {
		t.Fatal("qualified placement accepted")
	}
	for _, field := range []string{"zonegroup_id", "name", "realm_id", "placement_id", "storage_class", "expected_default_placement", "confirm_create"} {
		p := groupClassParams("")
		delete(p, field)
		if _, err := buildZonegroupStorageClass(p); err == nil {
			t.Fatalf("missing %s accepted", field)
		}
	}
}
