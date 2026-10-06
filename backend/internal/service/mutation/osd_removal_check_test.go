package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
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
	if err != nil || !reflect.DeepEqual(command.args, []string{"osd", "safe-to-destroy", "0", "2147483647", "--format", "json"}) {
		t.Fatalf("args=%v err=%v", command.args, err)
	}
}

func TestOSDRemovalCheckRunsReadOnly(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &osdSafetyExecutor{output: `{"safe_to_destroy":[0],"active":[],"missing_stats":[],"stored_pgs":[]}`}
	service.executor = runner
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

type osdSafetyExecutor struct {
	specs  []executor.CommandSpec
	output string
}

func (e *osdSafetyExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	return executor.CommandResult{Stdout: []byte(e.output)}, nil
}

func TestOSDRemovalReport(t *testing.T) {
	for _, tc := range []struct {
		raw         string
		valid, safe bool
	}{
		{`{"safe_to_destroy":[0],"active":[],"missing_stats":[],"stored_pgs":[]}`, true, true},
		{`{"safe_to_destroy":[],"active":[0],"missing_stats":[],"stored_pgs":[]}`, true, false},
		{`{"safe_to_destroy":[],"active":[0],"missing_stats":[],"stored_pgs":[0]}`, false, false},
		{`{"safe_to_destroy":[],"active":[],"missing_stats":[0],"stored_pgs":[]}`, true, false},
		{`{"safe_to_destroy":[],"active":[],"missing_stats":[],"stored_pgs":[]}`, true, false},
		{`{}`, false, false}, {`null`, false, false},
		{`{"safe_to_destroy":[null],"active":[],"missing_stats":[],"stored_pgs":[]}`, false, false},
		{`{"safe_to_destroy":[1],"active":[],"missing_stats":[],"stored_pgs":[]}`, false, false},
		{`{"safe_to_destroy":[0,0],"active":[],"missing_stats":[],"stored_pgs":[]}`, false, false},
		{`{"safe_to_destroy":[0],"active":[0],"missing_stats":[],"stored_pgs":[]}`, false, false},
	} {
		report, ok := osdRemovalReport([]byte(tc.raw), []string{"0"})
		if ok != tc.valid || ok && report["is_safe_to_destroy"] != tc.safe {
			t.Fatalf("raw=%s report=%v valid=%v", tc.raw, report, ok)
		}
	}
	service, _, id := newCephUserService(t)
	runner := &osdSafetyExecutor{output: `{"safe_to_destroy":[],"active":[0],"missing_stats":[],"stored_pgs":[]}`}
	service.executor = runner
	result, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.removal_check", Parameters: map[string]any{"osd_ids": []any{"0"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Details.(map[string]any)["check"].(map[string]any)["is_safe_to_destroy"] != false {
		t.Fatal("unsafe report treated as safe")
	}
	runner.output = `{}`
	if _, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.removal_check", Parameters: map[string]any{"osd_ids": []any{"0"}}}); err == nil {
		t.Fatal("invalid report accepted")
	}
}
