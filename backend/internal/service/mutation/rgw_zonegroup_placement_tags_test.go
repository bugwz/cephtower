package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"errors"
	"strings"
	"testing"
)

func groupTagsParams(def string, clear bool) map[string]any {
	p := groupClassParams(def)
	delete(p, "confirm_create")
	p["storage_class"] = "STANDARD"
	p["expected_tags"] = []any{"restricted"}
	p["tags"] = []any{"z", "a"}
	if clear {
		p["tags"] = []any{}
	}
	p["confirm_tags"] = true
	return p
}
func groupTagsFixture(def string, clear bool) *placementExecutor {
	r := groupClassFixture(def)
	tags := `["a","z"]`
	if clear {
		tags = `[]`
	}
	r.bodies["after"] = strings.Replace(r.bodies["before"], `["restricted"]`, tags, 1)
	if def == "" {
		r.bodies["after"] = strings.Replace(r.bodies["after"], `"default_placement":""`, `"default_placement":"p"`, 1)
	}
	return r
}
func TestZonegroupPlacementTags(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, def := range []string{"other", ""} {
		for _, clear := range []bool{false, true} {
			r := groupTagsFixture(def, clear)
			s.executor = r
			req := Request{ClusterID: cluster, Action: "rgw_zonegroup.placement_tags", Parameters: groupTagsParams(def, clear)}
			result, err := s.Execute(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			d := result.Details.(map[string]any)
			if d["tags_verified"] != true || d["period_published"] != false || d["default_placement_initialized"] != (def == "") {
				t.Fatal(d)
			}
			flag := "--tags=a,z"
			if clear {
				flag = "--tags-rm=restricted"
			}
			want := "zonegroup placement modify --zonegroup-id g --placement-id p --storage-class STANDARD --format json " + flag
			if len(r.calls) != 4 || strings.Join(r.calls[2].Args, " ") != want {
				t.Fatal(r.calls)
			}
			for _, stage := range []string{"before", "recheck", "add", "after"} {
				for _, mode := range []string{"exit", "error"} {
					broken := groupTagsFixture(def, clear)
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
					if broken.calls[len(broken.calls)-1].ID != "rgw_zonegroup.placement_tags."+stage {
						t.Fatal("continued after failure")
					}
				}
			}
		}
	}
	for _, scenario := range []string{"stale_tags", "missing_class", "concurrent", "tags_unchanged", "classes_changed", "default_changed"} {
		r := groupTagsFixture("other", true)
		switch scenario {
		case "stale_tags":
			r.bodies["before"] = strings.Replace(r.bodies["before"], "restricted", "new", 1)
		case "missing_class":
			r.bodies["before"] = strings.Replace(r.bodies["before"], "STANDARD", "COLD", 1)
		case "concurrent":
			r.bodies["recheck"] = strings.Replace(r.bodies["recheck"], "restricted", "new", 1)
		case "tags_unchanged":
			r.bodies["after"] = r.bodies["before"]
		case "classes_changed":
			r.bodies["after"] = strings.Replace(r.bodies["after"], "STANDARD", "COLD", 1)
		case "default_changed":
			r.bodies["after"] = strings.Replace(r.bodies["after"], `"default_placement":"other"`, `"default_placement":"p"`, 1)
		}
		s.executor = r
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.placement_tags", Parameters: groupTagsParams("other", true)})
		if err == nil {
			t.Fatal(scenario)
		}
		if scenario == "stale_tags" || scenario == "missing_class" || scenario == "concurrent" {
			for _, call := range r.calls {
				if call.Mutating {
					t.Fatal(scenario)
				}
			}
		}
	}
}
func TestZonegroupPlacementTagValidation(t *testing.T) {
	for _, key := range []string{"tags", "expected_tags"} {
		for _, bad := range []any{nil, []any{"a,b"}, []any{"a", "a"}, []any{""}, []any{"bad\n"}} {
			p := groupTagsParams("other", false)
			p[key] = bad
			if _, err := buildZonegroupPlacementTags(p); err == nil {
				t.Fatal(key, bad)
			}
		}
	}
	p := groupTagsParams("other", true)
	p["expected_tags"] = []any{}
	spec, err := buildZonegroupPlacementTags(p)
	if err != nil || strings.Contains(strings.Join(spec.args, " "), "--tags") {
		t.Fatal("empty no-op should omit tag flags")
	}
}

func TestZonegroupPlacementTagsPreserveNonstandardClassAndTier(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	r := groupTagsFixture("other", false)
	for _, stage := range []string{"before", "recheck", "after"} {
		r.bodies[stage] = strings.Replace(r.bodies[stage], `"storage_classes":["STANDARD"]`, `"storage_classes":["CLOUD"],"tier_targets":[{"key":"CLOUD","val":{"tier_type":"cloud-s3","s3":{"endpoint":"https://example.invalid"}}}]`, 1)
	}
	p := groupTagsParams("other", false)
	p["storage_class"] = "CLOUD"
	s.executor = r
	if _, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.placement_tags", Parameters: p}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(r.calls[2].Args, " "), "--storage-class CLOUD") {
		t.Fatal("must not insert STANDARD")
	}
}
