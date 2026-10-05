package mutation

import (
	"context"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func groupClassDeleteParams(realm string) map[string]any {
	return map[string]any{"zonegroup_id": "g", "name": "group", "realm_id": realm, "placement_id": "p", "storage_class": "COLD", "expected_default_placement": "p/COLD", "confirm_delete": true}
}
func groupClassDeleteFixture(realm string) *placementExecutor {
	r := placementFixture(realm)
	before := `{"id":"g","name":"group","realm_id":"` + realm + `","default_placement":"p/COLD","zones":[{"id":"z"}],"placement_targets":[{"name":"p","tags":["keep"],"storage_classes":["COLD","STANDARD"],"tier_targets":[{"key":"COLD","val":{"tier_type":"cloud-s3","s3":{"secret":"private"}}}]},{"name":"other","tags":[],"storage_classes":["STANDARD"]}]}`
	after := `{"id":"g","name":"group","realm_id":"` + realm + `","default_placement":"p","zones":[{"id":"z"}],"placement_targets":[{"name":"p","tags":["keep"],"storage_classes":["STANDARD"]},{"name":"other","tags":[],"storage_classes":["STANDARD"]}]}`
	r.bodies["before"], r.bodies["recheck"], r.bodies["after"], r.bodies["remove"] = before, before, after, `[]`
	return r
}
func TestZonegroupStorageClassDeletion(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, realm := range []string{"", "r"} {
		r := groupClassDeleteFixture(realm)
		s.executor = r
		req := Request{ClusterID: cluster, Action: "rgw_zonegroup.storage_class_delete", Parameters: groupClassDeleteParams(realm)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		d := result.Details.(map[string]any)
		if d["class_removal_verified"] != true || d["period_published"] != (realm != "") || d["objects_removed"] != false || d["zone_mappings_removed"] != false {
			t.Fatal(d)
		}
		writes := 0
		for _, call := range r.calls {
			if call.Mutating {
				writes++
			}
			if strings.HasSuffix(call.ID, ".remove") && strings.Join(call.Args, " ") != "zonegroup placement rm --zonegroup-id g --placement-id p --storage-class COLD --format json" {
				t.Fatal(call.Args)
			}
			for _, mode := range []string{"exit", "error"} {
				broken := groupClassDeleteFixture(realm)
				stage := strings.TrimPrefix(call.ID, "rgw_zonegroup.storage_class_delete.")
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
				if broken.calls[len(broken.calls)-1].ID != call.ID {
					t.Fatal("continued after failure")
				}
			}
		}
		want := 1
		if realm != "" {
			want = 2
		}
		if writes != want {
			t.Fatal("unexpected writes")
		}
	}
}
func TestZonegroupStorageClassDeleteGuards(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, scenario := range []string{"missing_class", "wrong_default", "concurrent", "tier_retained", "tags_changed", "member_changed", "default_changed"} {
		r := groupClassDeleteFixture("r")
		switch scenario {
		case "missing_class":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `["COLD","STANDARD"]`, `["STANDARD"]`, 1)
		case "wrong_default":
			r.bodies["before"] = strings.Replace(r.bodies["before"], `"default_placement":"p/COLD"`, `"default_placement":"other"`, 1)
		case "concurrent":
			r.bodies["recheck"] = strings.Replace(r.bodies["recheck"], "keep", "changed", 1)
		case "tier_retained":
			r.bodies["after"] = strings.Replace(r.bodies["after"], `"tags":["keep"]`, `"tags":["keep"],"tier_targets":[{"key":"COLD","val":{}}]`, 1)
		case "tags_changed":
			r.bodies["after"] = strings.Replace(r.bodies["after"], "keep", "changed", 1)
		case "member_changed":
			r.bodies["after"] = strings.Replace(r.bodies["after"], `"id":"z"`, `"id":"changed"`, 1)
		case "default_changed":
			r.bodies["after"] = strings.Replace(r.bodies["after"], `"default_placement":"p"`, `"default_placement":"other"`, 1)
		}
		s.executor = r
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.storage_class_delete", Parameters: groupClassDeleteParams("r")})
		if err == nil {
			t.Fatal(scenario)
		}
		for _, call := range r.calls {
			if strings.Contains(call.ID, "period") {
				t.Fatal("published after failed validation")
			}
			if call.Mutating && (scenario == "missing_class" || scenario == "wrong_default" || scenario == "concurrent") {
				t.Fatal("wrote before verification")
			}
		}
	}
}
func TestZonegroupClassDeletionNativeEmptyClassDefaults(t *testing.T) {
	for _, class := range []string{"COLD", "STANDARD"} {
		p := groupClassDeleteParams("")
		p["storage_class"] = class
		group := periodDocument([]byte(`{"default_placement":"p","placement_targets":[{"name":"p","tags":[],"storage_classes":["` + class + `"]}]}`))
		expected, ok := zonegroupStorageClassDeleteExpected(group, p)
		if class == "STANDARD" {
			if ok {
				t.Fatal("cannot remove sole STANDARD")
			}
			continue
		}
		if !ok || expected["placement_targets"].([]any)[0].(map[string]any)["storage_classes"].([]any)[0] != "STANDARD" {
			t.Fatal("missing native STANDARD recovery")
		}
	}
}
