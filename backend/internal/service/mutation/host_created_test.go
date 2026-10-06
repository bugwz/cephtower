package mutation

import (
	"context"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestHostCreationReadback(t *testing.T) {
	for _, tc := range []struct {
		raw                string
		maintenance, valid bool
	}{
		{`[{"hostname":"node1","status":"","labels":["osd"]}]`, false, true},
		{`[{"hostname":"node1","status":"maintenance","labels":["osd","extra"]}]`, true, true},
		{`[{"hostname":"node1","status":"","labels":["osd"]}]`, true, false},
		{`[{"hostname":"node1","status":"maintenance","labels":["osd"]}]`, false, false},
		{`[{"hostname":"node1","status":"","labels":[]}]`, false, false},
		{`[{"hostname":"node2","status":"","labels":["osd"]}]`, false, false},
		{`[{"hostname":"node1","labels":["osd"]}]`, false, false},
		{`[{"hostname":"node1","status":"","labels":null}]`, false, false},
		{`[{"hostname":"node1","status":"","labels":["osd"]},{"hostname":"node1","status":"","labels":["osd"]}]`, false, false},
		{`[]`, false, false}, {`null`, false, false},
	} {
		service, _, id := newCephUserService(t)
		runner := &hostRemovalExecutor{raw: tc.raw}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "host.create", ResourceKey: "host/node1", Parameters: map[string]any{"hostname": "node1", "labels": []any{"osd"}, "maintenance": tc.maintenance}})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
		if !tc.valid {
			var actionError *cephdomain.ActionError
			if !errors.As(err, &actionError) || actionError.Code != "post_check_failed" || actionError.Retryable {
				t.Fatal(err)
			}
		}
		if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating {
			t.Fatal(runner.specs)
		}
	}
}
