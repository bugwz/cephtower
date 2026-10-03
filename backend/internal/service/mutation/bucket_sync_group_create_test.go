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

func TestBucketSyncGroupCreation(t *testing.T) {
	old := `{"id":"old","status":"forbidden","data_flow":{},"pipes":[{"priority":9007199254740993}]}`
	added := `{"id":"new","status":"allowed","data_flow":{},"pipes":[]}`
	before := `{"groups":[` + old + `]}`
	after := `{"groups":[` + added + "," + old + `]}`
	for _, tc := range []struct {
		name, initial, actual string
		failAt, count         int
		code                  string
	}{
		{"create", before, after, 0, 3, ""},
		{"order irrelevant", before, `{"groups":[` + old + "," + added + `]}`, 0, 3, ""},
		{"empty policy", `{"groups":[]}`, `{"groups":[` + added + `]}`, 0, 3, ""},
		{"exists", after, after, 0, 1, "pre_check_failed"},
		{"unreadable", before, after, 1, 1, "pre_check_failed"},
		{"malformed", "broken", after, 0, 1, "pre_check_failed"},
		{"write uncertain", before, after, 2, 2, "command_failed"},
		{"post unreadable", before, after, 3, 3, "post_check_failed"},
		{"unchanged", before, before, 0, 3, "post_check_failed"},
		{"post malformed", before, "broken", 0, 3, "post_check_failed"},
		{"wrong status", before, strings.Replace(after, "allowed", "enabled", 1), 0, 3, "post_check_failed"},
		{"extra pipes", before, strings.Replace(after, `"pipes":[]`, `"pipes":[{"id":"unexpected"}]`, 1), 0, 3, "post_check_failed"},
		{"other policy changed", before, strings.Replace(after, "9007199254740993", "9007199254740992", 1), 0, 3, "post_check_failed"},
	} {
		for _, tenant := range []string{"", "team"} {
			t.Run(tc.name+"/"+tenant, func(t *testing.T) {
				service, _, clusterID := newCephUserService(t)
				runner := &syncGroupExecutor{before: tc.initial, after: tc.actual, failAt: tc.failAt}
				service.executor = runner
				bucketID := base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos"))
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_group_create", Parameters: map[string]any{"bucket_id": bucketID, "group_id": "new", "status": "allowed"}})
				if len(runner.calls) != tc.count {
					t.Fatalf("calls=%d want=%d", len(runner.calls), tc.count)
				}
				for index, call := range runner.calls {
					args := []string{"sync", "policy", "get", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					if index == 1 {
						args = []string{"sync", "group", "create", "--group-id", "new", "--status", "allowed", "--bucket", "photos", "--tenant", tenant, "--format", "json"}
					}
					if call.Binary != executor.BinaryRGWAdmin || call.Mutating != (index == 1) || !reflect.DeepEqual(call.Args, args) {
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
					t.Fatalf("unexpected failure: %v", err)
				}
			})
		}
	}
	for _, status := range []string{"enabled", "allowed", "forbidden"} {
		spec, err := build(Request{Action: "rgw_bucket.sync_group_create"}, map[string]any{"bucket_id": "AGJ1Y2tldA", "group_id": " new ", "status": status})
		if err != nil || spec.args[2] != "create" || spec.args[4] != " new " || spec.args[6] != status {
			t.Fatalf("invalid creation command: %+v %v", spec, err)
		}
	}
}
