package mutation

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type syncGroupExecutor struct {
	before, after string
	failAt        int
	calls         []executor.CommandSpec
}

func (e *syncGroupExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	if len(e.calls) == e.failAt {
		return executor.CommandResult{}, errors.New("native command unavailable")
	}
	if len(e.calls) == 1 {
		return executor.CommandResult{Stdout: []byte(e.before)}, nil
	}
	return executor.CommandResult{Stdout: []byte(e.after)}, nil
}

func TestBucketSyncGroupModification(t *testing.T) {
	before := `{"groups":[{"id":"target","status":"allowed","data_flow":{"directional":[]},"pipes":[{"id":"p","priority":9007199254740993}]},{"id":"other","status":"forbidden","data_flow":{},"pipes":[]}]}`
	after := strings.Replace(before, `"status":"allowed"`, `"status":"enabled"`, 1)
	for _, tc := range []struct {
		name, initial, actual, expected string
		failAt, count                   int
		code                            string
	}{
		{"success", before, after, "allowed", 0, 3, ""},
		{"stale", before, after, "enabled", 0, 1, "pre_check_failed"},
		{"missing", `{"groups":[]}`, after, "allowed", 0, 1, "pre_check_failed"},
		{"bad before", "broken", after, "allowed", 0, 1, "pre_check_failed"},
		{"read failure", before, after, "allowed", 1, 1, "pre_check_failed"},
		{"write uncertain", before, after, "allowed", 2, 2, "command_failed"},
		{"post failure", before, after, "allowed", 3, 3, "post_check_failed"},
		{"unchanged", before, before, "allowed", 0, 3, "post_check_failed"},
		{"bad after", before, "broken", "allowed", 0, 3, "post_check_failed"},
		{"pipe changed", before, strings.Replace(after, "9007199254740993", "9007199254740992", 1), "allowed", 0, 3, "post_check_failed"},
		{"other changed", before, strings.Replace(after, "forbidden", "enabled", 1), "allowed", 0, 3, "post_check_failed"},
	} {
		for _, tenant := range []string{"", "team"} {
			t.Run(tc.name+"/"+tenant, func(t *testing.T) {
				service, _, id := newCephUserService(t)
				runner := &syncGroupExecutor{before: tc.initial, after: tc.actual, failAt: tc.failAt}
				service.executor = runner
				bucketID := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos"))
				_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_bucket.sync_group", Parameters: map[string]any{"bucket_id": bucketID, "group_id": "target", "status": "enabled", "expected_status": tc.expected}})
				if len(runner.calls) != tc.count {
					t.Fatalf("calls=%d", len(runner.calls))
				}
				for i, call := range runner.calls {
					args := []string{"sync", "policy", "get", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					if i == 1 {
						args = []string{"sync", "group", "modify", "--group-id", "target", "--status", "enabled", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					}
					if call.Binary != executor.BinaryRGWAdmin || call.Mutating != (i == 1) || !reflect.DeepEqual(call.Args, args) {
						t.Fatalf("unsafe command: %+v", call)
					}
				}
				if tc.code == "" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				var actionError *cephdomain.ActionError
				if !errors.As(err, &actionError) || actionError.Code != tc.code || actionError.Retryable {
					t.Fatalf("unsafe failure: %v", err)
				}
			})
		}
	}
}

func TestBucketSyncGroupCommandValidation(t *testing.T) {
	exact, err := build(Request{Action: "rgw_bucket.sync_group"}, map[string]any{"bucket_id": "AGJ1Y2tldA", "group_id": " group ", "status": "allowed", "expected_status": "enabled"})
	if err != nil || exact.args[4] != " group " {
		t.Fatal("group identity whitespace was normalized")
	}
	for _, status := range []string{"enabled", "allowed", "forbidden"} {
		if _, err := build(Request{Action: "rgw_bucket.sync_group"}, map[string]any{"bucket_id": "AGJ1Y2tldA", "group_id": "group", "status": status, "expected_status": "allowed"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, change := range []map[string]any{{"bucket_id": "bad"}, {"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00-other"))}, {"group_id": "-bad"}, {"group_id": "a\nb"}, {"status": "future"}, {"expected_status": ""}} {
		params := map[string]any{"bucket_id": "AGJ1Y2tldA", "group_id": "group", "status": "enabled", "expected_status": "allowed"}
		for key, value := range change {
			params[key] = value
		}
		if _, err := build(Request{Action: "rgw_bucket.sync_group"}, params); err == nil {
			t.Fatalf("invalid request accepted: %#v", params)
		}
	}
}
