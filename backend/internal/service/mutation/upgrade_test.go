package mutation

import (
	"context"
	"reflect"
	"testing"
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
