package mutation

import (
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestBucketSyncGroupDeletion(t *testing.T) {
	group := `{"id":"g","status":"forbidden","data_flow":{},"pipes":[{"priority":9007199254740993}]}`
	other := `{"id":"other","status":"allowed","data_flow":{},"pipes":[]}`
	before := `{"groups":[` + group + "," + other + `]}`
	after := `{"groups":[` + other + `]}`
	for _, tc := range []struct {
		name, initial, expected, actual, code string
		failAt, count                         int
	}{
		{"remove", before, group, after, "", 0, 3},
		{"last group", `{"groups":[` + group + `]}`, group, `{"groups":[]}`, "", 0, 3},
		{"missing", after, group, after, "pre_check_failed", 0, 1},
		{"stale", before, other, after, "pre_check_failed", 0, 1},
		{"malformed expectation", before, "broken", after, "pre_check_failed", 0, 1},
		{"precision change", before, `{"id":"g","status":"forbidden","data_flow":{},"pipes":[{"priority":9007199254740992}]}`, after, "pre_check_failed", 0, 1},
		{"pre read", before, group, after, "pre_check_failed", 1, 1},
		{"write", before, group, after, "command_failed", 2, 2},
		{"post read", before, group, after, "post_check_failed", 3, 3},
		{"still exists", before, group, before, "post_check_failed", 0, 3},
		{"other removed", before, group, `{"groups":[]}`, "post_check_failed", 0, 3},
		{"malformed post", before, group, "broken", "post_check_failed", 0, 3},
	} {
		for _, tenant := range []string{"", "team"} {
			t.Run(tc.name+"/"+tenant, func(t *testing.T) {
				s, _, clusterID := newCephUserService(t)
				runner := &syncGroupExecutor{before: tc.initial, after: tc.actual, failAt: tc.failAt}
				s.executor = runner
				_, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_group_delete", Parameters: map[string]any{
					"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "expected_group": tc.expected,
				}})
				if len(runner.calls) != tc.count {
					t.Fatalf("calls=%d want=%d", len(runner.calls), tc.count)
				}
				for i, call := range runner.calls {
					args := []string{"sync", "policy", "get", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					if i == 1 {
						args = []string{"sync", "group", "remove", "--group-id", "g", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					}
					if call.Binary != executor.BinaryRGWAdmin || call.Mutating != (i == 1) || !reflect.DeepEqual(call.Args, args) {
						t.Fatalf("wrong command: %+v", call)
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
					t.Fatalf("unexpected error: %v", err)
				}
			})
		}
	}
}
