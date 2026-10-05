package mutation

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func groupPlacementParams(def string) map[string]any {
	return map[string]any{"zonegroup_id": "g", "name": "group", "realm_id": "r", "placement_id": "new", "tags": []any{"z", "a"}, "expected_default_placement": def, "confirm_create": true}
}
func groupPlacementFixture(def string) *placementExecutor {
	old := ""
	if def != "" {
		old = `{"name":"existing","tags":["keep"],"storage_classes":["STANDARD"]}`
	}
	before := `{"id":"g","name":"group","realm_id":"r","default_placement":"` + def + `","zones":[{"id":"z"}],"placement_targets":[` + old + `]}`
	targets := old
	if targets != "" {
		targets += ","
	}
	targets += `{"name":"new","tags":["a","z"],"storage_classes":["STANDARD"]}`
	after := strings.Replace(before, `"placement_targets":[`+old+`]`, `"placement_targets":[`+targets+`]`, 1)
	if def == "" {
		after = strings.Replace(after, `"default_placement":""`, `"default_placement":"new"`, 1)
	}
	return &placementExecutor{bodies: map[string]string{"before": before, "recheck": before, "add": `[]`, "after": after}}
}
func TestZonegroupPlacementCreation(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, def := range []string{"", "existing"} {
		runner := groupPlacementFixture(def)
		s.executor = runner
		req := Request{ClusterID: cluster, Action: "rgw_zonegroup.placement_create", Parameters: groupPlacementParams(def)}
		result, err := s.Execute(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		d := result.Details.(map[string]any)
		if d["placement_created"] != true || d["storage_class"] != "STANDARD" || d["default_placement_initialized"] != (def == "") || d["period_published"] != false {
			t.Fatal("incorrect creation result")
		}
		if len(runner.calls) != 4 {
			t.Fatal("unexpected calls")
		}
		want := []string{"zonegroup", "placement", "add", "--zonegroup-id", "g", "--placement-id", "new", "--storage-class", "STANDARD", "--format", "json", "--tags=z,a"}
		if !reflect.DeepEqual(runner.calls[2].Args, want) {
			t.Fatalf("incorrect command: %v", runner.calls[2].Args)
		}
		for i, c := range runner.calls {
			if c.Mutating != (i == 2) {
				t.Fatal("unexpected mutation")
			}
		}
		for _, stage := range []string{"before", "recheck", "add", "after"} {
			for _, mode := range []string{"exit", "error"} {
				broken := groupPlacementFixture(def)
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
				if broken.calls[len(broken.calls)-1].ID != req.Action+"."+stage {
					t.Fatal("continued after error")
				}
			}
		}
	}
}
func TestZonegroupPlacementRejectsOverwriteAndUnexpectedChanges(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	for _, tc := range []struct{ stage, from, to string }{
		{"before", `"name":"existing"`, `"name":"new"`},
		{"before", `"default_placement":"existing"`, `"default_placement":"changed"`},
		{"recheck", `"keep"`, `"changed"`},
		{"after", `"tags":["a","z"]`, `"tags":[]`},
		{"after", `"keep"`, `"changed"`},
		{"after", `"default_placement":"existing"`, `"default_placement":"new"`},
	} {
		runner := groupPlacementFixture("existing")
		runner.bodies[tc.stage] = strings.Replace(runner.bodies[tc.stage], tc.from, tc.to, 1)
		s.executor = runner
		_, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_zonegroup.placement_create", Parameters: groupPlacementParams("existing")})
		if err == nil {
			t.Fatalf("accepted drift %s", tc.stage)
		}
		if tc.stage != "after" {
			for _, c := range runner.calls {
				if c.Mutating {
					t.Fatal("wrote despite preflight drift")
				}
			}
		}
	}
	for _, tags := range []any{nil, []any{"duplicate", "duplicate"}, []any{"a,b"}, []any{""}, []any{"control\n"}} {
		p := groupPlacementParams("")
		p["tags"] = tags
		if _, err := buildZonegroupPlacementCreate(p); err == nil {
			t.Fatal("invalid tags accepted")
		}
	}
	p := groupPlacementParams("")
	p["tags"] = []any{}
	spec, err := buildZonegroupPlacementCreate(p)
	if err != nil || strings.Contains(strings.Join(spec.args, " "), "--tags") {
		t.Fatal("empty tags should omit flag")
	}
	before := periodDocument([]byte(groupPlacementFixture("").bodies["before"]))
	expected, ok := zonegroupPlacementCreateExpected(before, p)
	if !ok || len(expected["placement_targets"].([]any)[0].(map[string]any)["tags"].([]any)) != 0 {
		t.Fatal("empty tags not preserved")
	}
}
