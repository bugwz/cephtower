package mutation

import (
	"context"
	"reflect"
	"testing"
)

func TestOSDRemovalCheckRejectsAmbiguousIDs(t *testing.T) {
	for _, ids := range [][]string{nil, {}, {""}, {"--help"}, {"-1"}, {"+1"}, {"01"}, {" 1"}, {"1 "}, {"1.0"}, {"1", "1"}, {"2147483648"}, {"1;id"}} {
		values := make([]any, len(ids))
		for i, id := range ids {
			values[i] = id
		}
		if _, err := build(Request{Action: "osd.removal_check"}, map[string]any{"osd_ids": values}); err == nil {
			t.Fatalf("accepted %v", ids)
		}
	}
	command, err := build(Request{Action: "osd.removal_check"}, map[string]any{"osd_ids": []any{"0", "2147483647"}})
	if err != nil || !reflect.DeepEqual(command.args, []string{"osd", "safe-to-destroy", "0", "2147483647"}) {
		t.Fatalf("args=%v err=%v", command.args, err)
	}
}

func TestOSDRemovalCheckRunsReadOnly(t *testing.T) {
	service, runner, id := newCephUserService(t)
	_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.removal_check", ResourceKey: "osd/removal-check", Parameters: map[string]any{"osd_ids": []any{"0"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 1 || runner.specs[0].Mutating {
		t.Fatalf("specs=%+v", runner.specs)
	}
	count := len(runner.specs)
	if _, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.removal_check", Parameters: map[string]any{"osd_ids": []any{"--help"}}}); err == nil {
		t.Fatal("invalid ID accepted")
	}
	if len(runner.specs) != count {
		t.Fatal("invalid ID reached executor")
	}
}
