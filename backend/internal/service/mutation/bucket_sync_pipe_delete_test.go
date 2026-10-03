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

func TestBucketSyncPipeDeletion(t *testing.T) {
	target := `{"id":" target ","source":{"bucket":"*","zones":["*"]},"dest":{"bucket":"team/photos:marker","zones":["Zone B"]},"params":{"priority":9007199254740993}}`
	other := `{"id":"other","source":{"bucket":"*"},"dest":{"bucket":"*"},"params":{"mode":"user"}}`
	group := func(pipes string) string {
		return `{"id":"g","status":"forbidden","data_flow":{"symmetrical":[{"id":"flow","zones":["Zone A","Zone B"]}]},"pipes":[` + pipes + `]}`
	}
	before := group(target + "," + other)
	after := group(other)
	policy := func(g string) string {
		return `{"groups":[` + g + `,{"id":"other","status":"allowed","data_flow":{},"pipes":[]}]}`
	}
	for _, tc := range []struct {
		name, initial, expected, actual, code string
		fail, count                           int
	}{
		{"remove", before, before, policy(after), "", 0, 3},
		{"last pipe", group(target), group(target), policy(group("")), "", 0, 3},
		{"missing", after, after, policy(after), "pre_check_failed", 0, 1},
		{"duplicate", group(target + "," + target), group(target + "," + target), policy(after), "pre_check_failed", 0, 1},
		{"malformed pipe", group("null"), group("null"), policy(after), "pre_check_failed", 0, 1},
		{"stale", before, after, policy(after), "pre_check_failed", 0, 1},
		{"malformed expectation", before, "broken", policy(after), "pre_check_failed", 0, 1},
		{"precision changed", before, strings.Replace(before, "9007199254740993", "9007199254740992", 1), policy(after), "pre_check_failed", 0, 1},
		{"pre read", before, before, policy(after), "pre_check_failed", 1, 1},
		{"write", before, before, policy(after), "command_failed", 2, 2},
		{"post read", before, before, policy(after), "post_check_failed", 3, 3},
		{"unchanged", before, before, policy(before), "post_check_failed", 0, 3},
		{"other pipe lost", before, before, policy(group("")), "post_check_failed", 0, 3},
		{"flow changed", before, before, strings.Replace(policy(after), "Zone A", "Changed", 1), "post_check_failed", 0, 3},
		{"group changed", before, before, strings.Replace(policy(after), "forbidden", "enabled", 1), "post_check_failed", 0, 3},
		{"other group lost", before, before, `{"groups":[` + after + `]}`, "post_check_failed", 0, 3},
	} {
		for _, tenant := range []string{"", "team"} {
			t.Run(tc.name+"/"+tenant, func(t *testing.T) {
				s, _, clusterID := newCephUserService(t)
				runner := &syncGroupExecutor{before: policy(tc.initial), after: tc.actual, failAt: tc.fail}
				s.executor = runner
				p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "pipe_id": " target ", "expected_group": tc.expected}
				_, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_pipe_delete", Parameters: p})
				if len(runner.calls) != tc.count {
					t.Fatalf("calls=%d expected=%d err=%v", len(runner.calls), tc.count, err)
				}
				for i, call := range runner.calls {
					args := []string{"sync", "policy", "get"}
					if i == 1 {
						args = []string{"sync", "group", "pipe", "remove", "--group-id", "g", "--pipe-id", " target "}
					}
					args = append(args, "--bucket", "photos", "--tenant", tenant, "--format", "json")
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
				var e *cephdomain.ActionError
				if !errors.As(err, &e) || e.Code != tc.code || e.Retryable {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		}
	}
	for _, id := range []string{"", "-bad", "a\nb", strings.Repeat("x", 513)} {
		if _, err := bucketSyncPipeDeleteArgs(map[string]any{"pipe_id": id, "expected_group": "{}"}); err == nil {
			t.Fatalf("accepted invalid ID: %q", id)
		}
	}
}
