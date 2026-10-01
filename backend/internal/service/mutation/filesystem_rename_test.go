package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestFilesystemRenameCommand(t *testing.T) {
	p := map[string]any{"new_name": "renamed", "confirmed": true}
	cmd, err := build(Request{Action: "filesystem.rename", ResourceKey: "filesystem/original"}, p)
	if err != nil || !reflect.DeepEqual(cmd.args, []string{"fs", "volume", "rename", "original", "renamed", "--yes-i-really-mean-it"}) || !reflect.DeepEqual(cmd.check, []string{"fs", "volume", "ls", "--format", "json"}) || !Supports("filesystem.rename") {
		t.Fatalf("command = %+v, err = %v", cmd, err)
	}
	for _, parameters := range []map[string]any{
		{"new_name": "renamed"}, {"new_name": "renamed", "confirmed": false},
		{"new_name": "renamed", "confirmed": "true"}, {"new_name": "original", "confirmed": true},
		{"new_name": "--help", "confirmed": true}, {"new_name": "bad/name", "confirmed": true},
		{"new_name": "bad name", "confirmed": true}, {"new_name": "", "confirmed": true},
	} {
		if _, err := build(Request{Action: "filesystem.rename", ResourceKey: "filesystem/original"}, parameters); err == nil {
			t.Fatalf("accepted invalid rename: %v", parameters)
		}
	}
	for _, key := range []string{"filesystem/team/original", "filesystem/--help", "original", "filesystem/"} {
		if _, err := build(Request{Action: "filesystem.rename", ResourceKey: key}, p); err == nil {
			t.Fatalf("accepted invalid resource key %q", key)
		}
	}
	if _, err := build(Request{Action: "filesystem.rename", ResourceKey: "filesystem/123-existing"}, p); err != nil {
		t.Fatalf("rejected native existing name: %v", err)
	}
}

func TestFilesystemRenameReadback(t *testing.T) {
	for _, tt := range []struct {
		data string
		want bool
	}{
		{`[{"name":"renamed"},{"name":"unrelated"}]`, true},
		{`[{"name":"original"},{"name":"renamed"}]`, false},
		{`[{"name":"unrelated"}]`, false}, {`[]`, false}, {`null`, false},
		{`{}`, false}, {`[{"name":"renamed"},{}]`, false},
		{`[{"name":"renamed"},{"name":"renamed"}]`, false}, {`invalid`, false},
	} {
		if got := filesystemRenameMatches("original", "renamed", []byte(tt.data)); got != tt.want {
			t.Fatalf("readback %s = %v, want %v", tt.data, got, tt.want)
		}
	}
}

func TestFilesystemRenameExecution(t *testing.T) {
	for _, tt := range []struct {
		name, data, failID string
		wantSuccess        bool
	}{
		{"success", `[{"name":"renamed"}]`, "", true},
		{"old still present", `[{"name":"original"},{"name":"renamed"}]`, "", false},
		{"new missing", `[]`, "", false},
		{"check fails", `[{"name":"renamed"}]`, "filesystem.rename.post_check", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"filesystem.rename.post_check": tt.data, "filesystem.rename": "volume renamed; multiple data pools were not renamed"}, failID: tt.failID}
			service.executor = runner
			result, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "filesystem.rename", ResourceKey: "filesystem/original", Parameters: map[string]any{"new_name": "renamed", "confirmed": true}})
			if tt.wantSuccess {
				if err != nil || result.Details.(map[string]any)["native_output"] != runner.outputs["filesystem.rename"] {
					t.Fatalf("result = %+v, err = %v", result, err)
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" {
					t.Fatalf("error = %v", err)
				}
			}
			if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating {
				t.Fatalf("unexpected execution chain: %+v", runner.specs)
			}
		})
	}
}
