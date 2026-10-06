package mutation

import (
	"context"
	"testing"
)

func TestHostCreateRejectsLossyLabelsBeforeExecution(t *testing.T) {
	for _, labels := range []any{
		[]any{"osd,mon"}, []any{""}, []any{" osd"}, []any{"osd "},
		[]any{"--maintenance"}, []any{"osd", "osd"}, []any{42}, "osd",
	} {
		service, _, id := newCephUserService(t)
		runner := &hostRemovalExecutor{}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.create", ResourceKey: "host/node1", Parameters: map[string]any{"hostname": "node1", "labels": labels}})
		if err == nil || len(runner.specs) != 0 {
			t.Fatalf("labels=%v err=%v commands=%v", labels, err, runner.specs)
		}
	}
}
