package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestBucketSyncPipeUpdate(t *testing.T) {
	group := `{"id":"g","status":"enabled","data_flow":{"symmetrical":[{"id":"f","zones":["A"]}]},"pipes":[{"id":"p","source":{"bucket":"old","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"user","user":"old-user","priority":9007199254740993,"source":{"filter":{"prefix":"keep","tags":[{"key":"x","value":"y"}]}},"dest":{"storage_class":"COLD"}}}]}`
	policy := func(g string) string {
		return `{"groups":[` + g + `,{"id":"other","status":"forbidden","data_flow":{},"pipes":[]}]}`
	}
	for _, mode := range []string{"system", "user"} {
		changed := strings.Replace(group, `"bucket":"old"`, `"bucket":"team/photos:instance"`, 1)
		changed = strings.Replace(changed, `"mode":"user"`, `"mode":"`+mode+`"`, 1)
		if mode == "user" {
			changed = strings.Replace(changed, `"user":"old-user"`, `"user":"team$new-user"`, 1)
		}
		for _, tenant := range []string{"", "team"} {
			for _, tc := range []struct {
				name, before, after, expected, code string
				fail, count                         int
			}{
				{"success", group, policy(changed), group, "", 0, 3},
				{"stale", group, policy(changed), changed, "pre_check_failed", 0, 1},
				{"missing", strings.Replace(group, `"id":"p"`, `"id":"other"`, 1), "", "", "pre_check_failed", 0, 1},
				{"read", group, "", group, "pre_check_failed", 1, 1},
				{"write", group, "", group, "command_failed", 2, 2},
				{"post", group, "", group, "post_check_failed", 3, 3},
				{"unchanged result", group, policy(group), group, "post_check_failed", 0, 3},
				{"changed zones", group, policy(strings.Replace(changed, `"zones":["A"]`, `"zones":["B"]`, -1)), group, "post_check_failed", 0, 3},
				{"changed advanced param", group, policy(strings.Replace(changed, "9007199254740993", "9007199254740992", 1)), group, "post_check_failed", 0, 3},
				{"lost user", group, policy(strings.Replace(changed, `"user":"old-user",`, "", 1)), group, "", 0, 3},
			} {
				t.Run(mode+"/"+tenant+"/"+tc.name, func(t *testing.T) {
					if tc.name == "lost user" && mode == "system" {
						tc.code = "post_check_failed"
					}
					expected := tc.expected
					if expected == "" {
						expected = tc.before
					}
					service, _, id := newCephUserService(t)
					runner := &syncGroupExecutor{before: policy(tc.before), after: tc.after, failAt: tc.fail}
					service.executor = runner
					p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00bucket")), "group_id": "g", "pipe_id": "p", "expected_group": expected, "source_tenant": "team", "source_bucket": "photos", "source_bucket_id": "instance", "dest_bucket": "*", "mode": mode}
					if mode == "user" {
						p["user"] = "team$new-user"
					}
					_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_bucket.sync_pipe_update", Parameters: p})
					if tc.code == "" {
						if err != nil {
							t.Fatal(err)
						}
					} else {
						var ae *cephdomain.ActionError
						if !errors.As(err, &ae) || ae.Code != tc.code || ae.Retryable {
							t.Fatalf("error %v", err)
						}
					}
					if len(runner.calls) != tc.count {
						t.Fatalf("calls: %v", runner.calls)
					}
					if len(runner.calls) >= 2 {
						call := runner.calls[1]
						want := []string{"sync", "group", "pipe", "modify", "--group-id", "g", "--pipe-id", "p", "--source-tenant", "team", "--source-bucket", "photos", "--source-bucket-id", "instance", "--dest-tenant", "", "--dest-bucket", "*", "--dest-bucket-id", "*", "--mode", mode}
						if mode == "user" {
							want = append(want, "--uid", "team$new-user")
						}
						want = append(want, "--bucket", "bucket", "--tenant", tenant, "--format", "json")
						if !call.Mutating || call.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(call.Args, want) {
							t.Fatalf("command: %+v want %v", call, want)
						}
					}
				})
			}
		}
	}
}

func TestBucketSyncPipeUpdateValidation(t *testing.T) {
	p := map[string]any{"pipe_id": "p", "expected_group": "{}", "source_bucket": "*", "dest_bucket": "*", "mode": "system"}
	validPipe := `{"id":"p","source":{"bucket":"*","zones":["A"]},"dest":{"bucket":"*","zones":["*"]},"params":{"mode":"system"}}`
	for _, pipes := range []string{validPipe, validPipe + "," + validPipe} {
		_, groups, _ := bucketSyncPolicyDocument([]byte(`{"groups":[{"id":"g","status":"enabled","data_flow":{},"pipes":[` + pipes + `]}]}`))
		if err := updateBucketSyncPipe(groups["g"], p); err == nil {
			t.Fatal("accepted unchanged or duplicate pipe")
		}
	}
	for _, change := range []map[string]any{{"pipe_id": "-bad"}, {"mode": "unknown"}, {"user": "unexpected"}, {"source_bucket": ""}, {"source_bucket": "a/b"}, {"source_tenant": "a:b"}, {"dest_bucket_id": "a b"}, {"mode": "user", "user": ""}, {"mode": "user", "user": "a$$"}} {
		copy := map[string]any{}
		for k, v := range p {
			copy[k] = v
		}
		for k, v := range change {
			copy[k] = v
		}
		if _, err := bucketSyncPipeUpdateArgs(copy); err == nil {
			t.Fatalf("accepted %v", change)
		}
	}
	for _, pipe := range []string{`null`, `{"id":"p"}`, `{"id":"p","params":{},"source":null}`, `{"id":"other"}`} {
		_, groups, ok := bucketSyncPolicyDocument([]byte(`{"groups":[{"id":"g","status":"enabled","data_flow":{},"pipes":[` + pipe + `]}]}`))
		if !ok {
			t.Fatal("fixture")
		}
		if err := updateBucketSyncPipe(groups["g"], p); err == nil {
			t.Fatalf("accepted %s", pipe)
		}
	}
}
