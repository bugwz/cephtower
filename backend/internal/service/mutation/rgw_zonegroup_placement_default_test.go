package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"strings"
	"testing"
)

func defaultPlacementFixture(class string) *placementExecutor {
	r := groupClassFixture("other")
	before := strings.Replace(r.bodies["before"], `["STANDARD"]`, `["COLD","STANDARD"]`, 1)
	def := "p"
	if class != "STANDARD" {
		def += "/" + class
	}
	r.bodies["before"], r.bodies["recheck"] = before, before
	r.bodies["after"] = strings.Replace(before, `"default_placement":"other"`, `"default_placement":"`+def+`"`, 1)
	return r
}
func TestZonegroupPlacementDefault(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, class := range []string{"STANDARD", "COLD"} {
		p := groupClassParams("other")
		delete(p, "confirm_create")
		p["confirm_default"] = true
		p["storage_class"] = class
		req := Request{ClusterID: cluster, Action: "rgw_zonegroup.placement_default", Parameters: p}
		r := defaultPlacementFixture(class)
		s.executor = r
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		d := result.Details.(map[string]any)
		if d["default_placement_verified"] != true || d["period_published"] != false || d["data_migrated"] != false {
			t.Fatal(d)
		}
		if len(r.calls) != 4 || strings.Join(r.calls[2].Args, " ") != "zonegroup placement default --zonegroup-id g --placement-id p --storage-class "+class+" --format json" {
			t.Fatal(r.calls)
		}
		for _, stage := range []string{"before", "recheck", "add", "after"} {
			for _, mode := range []string{"exit", "error"} {
				broken := defaultPlacementFixture(class)
				if mode == "exit" {
					broken.codeStage = stage
				} else {
					broken.fail = stage
				}
				s.executor = broken
				_, err := s.Execute(context.Background(), req)
				var ae *cephdomain.ActionError
				if !errors.As(err, &ae) || ae.Retryable || strings.Contains(err.Error(), "private") {
					t.Fatal(err)
				}
				if broken.calls[len(broken.calls)-1].ID != "rgw_zonegroup.placement_default."+stage {
					t.Fatal("continued after failure")
				}
			}
		}
		for _, scenario := range []string{"missing_class", "missing_target", "changed_default", "concurrent_change", "tags_changed", "classes_changed"} {
			broken := defaultPlacementFixture(class)
			switch scenario {
			case "missing_class":
				broken.bodies["before"] = strings.ReplaceAll(broken.bodies["before"], `["COLD","STANDARD"]`, `["OTHER"]`)
			case "missing_target":
				broken.bodies["before"] = strings.ReplaceAll(broken.bodies["before"], `"name":"p"`, `"name":"missing"`)
			case "changed_default":
				broken.bodies["before"] = strings.ReplaceAll(broken.bodies["before"], `"default_placement":"other"`, `"default_placement":"changed"`)
			case "concurrent_change":
				broken.bodies["recheck"] = strings.ReplaceAll(broken.bodies["recheck"], "restricted", "changed")
			case "tags_changed":
				broken.bodies["after"] = strings.ReplaceAll(broken.bodies["after"], "restricted", "changed")
			case "classes_changed":
				broken.bodies["after"] = strings.ReplaceAll(broken.bodies["after"], `["COLD","STANDARD"]`, `["STANDARD"]`)
			}
			s.executor = broken
			if _, err := s.Execute(context.Background(), req); err == nil {
				t.Fatal(scenario)
			}
			if scenario != "tags_changed" && scenario != "classes_changed" {
				for _, call := range broken.calls {
					if call.Mutating {
						t.Fatal(scenario)
					}
				}
			}
		}
	}
}
