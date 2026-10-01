package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestUpgradeControlReadback(t *testing.T) {
	s, _, id := newCephUserService(t)
	for action, good := range map[string]string{
		"pause":  `{"in_progress":true,"is_paused":true}`,
		"resume": `{"in_progress":true,"is_paused":false}`,
		"stop":   `{"in_progress":false,"is_paused":false}`,
	} {
		r := Request{ClusterID: id, Action: "upgrade.action", ResourceKey: "upgrade/action", Parameters: map[string]any{"action": action}}
		e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "accepted", r.Action + ".post_check": good}}
		s.executor = e
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(action, err)
		}
		if len(e.specs) != 2 || !reflect.DeepEqual(e.specs[0].Args, []string{"orch", "upgrade", action}) || !reflect.DeepEqual(e.specs[1].Args, []string{"orch", "upgrade", "status", "--format", "json"}) {
			t.Fatal(e.specs)
		}
		for _, bad := range []string{`null`, `{}`, `[]`, `{"in_progress":false}`, `{"in_progress":"true","is_paused":false}`, `{"in_progress":false,"is_paused":true}`, good + ` {}`} {
			e.outputs[r.Action+".post_check"] = bad
			if _, err := s.Execute(context.Background(), r); err == nil {
				t.Fatalf("%s accepted %s", action, bad)
			}
		}
	}
}

func TestUpgradeCheckUsesVersionOption(t *testing.T) {
	p := map[string]any{"version": "20.2.2"}
	command, err := build(Request{Action: "upgrade.check", ResourceKey: "upgrade/check", Parameters: p}, p)
	if err != nil {
		t.Fatal(err)
	}
	// The first positional argument is image, not ceph_version, in OrchestratorCli.
	want := []string{"orch", "upgrade", "check", "--ceph-version", "20.2.2", "--format", "json"}
	if !reflect.DeepEqual(command.args, want) {
		t.Fatalf("got %v, want %v", command.args, want)
	}
	if _, err := build(Request{Action: "upgrade.check", ResourceKey: "upgrade/check"}, map[string]any{}); err == nil {
		t.Fatal("missing version accepted")
	}
}

func TestUpgradeCheckReturnsReport(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "upgrade.check", ResourceKey: "upgrade/check", Parameters: map[string]any{"version": "20.2.2"}}
	good := `{"target_name":"ceph:v20.2.2","target_id":"sha256:abc","target_version":"20.2.2","needs_update":{"mon.a":{"current_name":null,"current_id":null,"current_version":null,"password":"do-not-copy"}},"up_to_date":[],"non_ceph_image_daemons":["prometheus.a"],"password":"do-not-copy"}`
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: good}}
	s.executor = e
	result, err := s.Execute(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	report, ok := result.Details.(map[string]any)["check"].(map[string]any)
	if !ok || report["target_version"] != "20.2.2" || report["password"] != nil || report["needs_update"].(map[string]any)["mon.a"].(map[string]any)["password"] != nil {
		t.Fatal(result)
	}
	for _, bad := range []string{"Incompatible upgrade: downgrade is not supported", "Unable to extract ceph version", `null`, `{}`, good + `{}`, `{"target_name":1}`} {
		e.outputs[r.Action] = bad
		if _, err := s.Execute(context.Background(), r); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}

func TestUpgradeStartChecksTarget(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "upgrade.action", ResourceKey: "upgrade/action", Parameters: map[string]any{"action": "start", "version": "20.2.2"}}
	check := `{"target_name":"ceph:v20.2.2","target_id":"abc","target_version":"20.2.2","target_digest":"ceph@sha256:abc","needs_update":{},"up_to_date":[],"non_ceph_image_daemons":[]}`
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action + ".pre_check": check, r.Action: "Initiating upgrade"}}
	s.executor = e
	for _, image := range []string{"ceph:v20.2.2", "ceph@sha256:abc"} {
		e.specs = nil
		e.outputs[r.Action+".post_check"] = `{"in_progress":true,"is_paused":false,"target_image":"` + image + `"}`
		if _, err := s.Execute(context.Background(), r); err != nil {
			t.Fatal(err)
		}
		if len(e.specs) != 3 || !reflect.DeepEqual(e.specs[0].Args, []string{"orch", "upgrade", "check", "--ceph-version", "20.2.2", "--format", "json"}) {
			t.Fatal(e.specs)
		}
	}
	for _, raw := range []string{`{}`, `{"in_progress":true,"is_paused":false,"target_image":"other"}`, `{"in_progress":true,"is_paused":true,"target_image":"ceph:v20.2.2"}`} {
		e.outputs[r.Action+".post_check"] = raw
		if _, err := s.Execute(context.Background(), r); err == nil {
			t.Fatal("unverified target accepted")
		}
	}
	e.specs = nil
	e.failID = r.Action + ".post_check"
	_, err := s.Execute(context.Background(), r)
	var actionError *cephdomain.ActionError
	if !errors.As(err, &actionError) || actionError.Retryable || actionError.Code != "post_check_failed" {
		t.Fatalf("uncertain start may be retried: %v", err)
	}
	e.failID = ""
	e.specs = nil
	e.outputs[r.Action+".pre_check"] = "Incompatible upgrade"
	if _, err := s.Execute(context.Background(), r); err == nil || len(e.specs) != 1 {
		t.Fatal("started despite failed check", e.specs)
	}
}
