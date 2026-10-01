package mutation

import (
	"context"
	"errors"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestCrushRuleCreationReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	p := map[string]any{"name": "ssd-rule", "root": "default", "device_class": "ssd", "failure_domain": "host"}
	r := Request{ClusterID: id, Action: "crush_rule.create", ResourceKey: "crush-rule", Parameters: p}
	good := `{"rule_name":"ssd-rule","type":1,"steps":[{"op":"take","item_name":"default~ssd"},{"op":"chooseleaf_firstn","num":0,"type":"host"},{"op":"emit"}]}`
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "already exists", r.Action + ".post_check": good}}
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`null`, `{}`, good + `{}`, strings.Replace(good, "default~ssd", "other~ssd", 1), strings.Replace(good, "host", "rack", 1), strings.Replace(good, `"type":1`, `"type":3`, 1), strings.Replace(good, `"num":0`, `"num":2`, 1), strings.Replace(good, "ssd-rule", "different", 1)} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
	p["failure_domain"] = "osd"
	if !crushRuleCreated(p, []byte(strings.Replace(strings.Replace(good, "chooseleaf_firstn", "choose_firstn", 1), "host", "osd", 1))) {
		t.Fatal("OSD failure domain rejected")
	}
}

func TestCrushRuleRenameReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "crush_rule.update", ResourceKey: "crush-rule/old", Parameters: map[string]any{"name": "new"}}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "renamed", r.Action + ".post_check": `["new","other"]`}}
	s.executor = e
	if _, err := s.Execute(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{`[]`, `null`, `{}`, `[null]`, `["old","new"]`, `["new","new"]`, `["new"] {}`} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
	if !crushRuleRenamed("same", "same", []byte(`["same"]`)) {
		t.Fatal("same-name operation rejected")
	}
}

func TestCrushRuleDeletionReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "crush_rule.delete", ResourceKey: "crush-rule/remove-me"}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "removed"}}
	s.executor = e
	for _, good := range []string{`[]`, `["keep-me"]`} {
		e.outputs[r.Action+".post_check"] = good
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(good, err)
		}
	}
	for _, bad := range []string{`null`, `{}`, `[null]`, `[""]`, `["remove-me"]`, `[] {}`} {
		e.outputs[r.Action+".post_check"] = bad
		_, err := s.Execute(context.Background(), r)
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(bad, err)
		}
	}
}
