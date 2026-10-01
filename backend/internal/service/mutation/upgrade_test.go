package mutation

import (
	"reflect"
	"testing"
)

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
