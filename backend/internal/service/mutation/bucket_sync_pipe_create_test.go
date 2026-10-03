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

func pipeCreateParameters() map[string]any {
	return map[string]any{"group_id": "g", "pipe_id": " new ", "expected_group": "{}", "source_zones": []any{"b", "a"}, "dest_zones": []any{"*"}, "source_tenant": "team", "source_bucket": "photos", "source_bucket_id": "marker", "dest_tenant": "*", "dest_bucket": "*", "mode": "system"}
}

func TestBucketSyncPipeCreation(t *testing.T) {
	other := `{"id":"other","params":{"priority":9007199254740993}}`
	group := func(pipes string) string {
		return `{"id":"g","status":"allowed","data_flow":{"directional":[{"source_zone":"Zeta","dest_zone":"Alpha"}]},"pipes":[` + pipes + `]}`
	}
	policy := func(g string) string {
		return `{"groups":[` + g + `,{"id":"other-group","status":"forbidden","data_flow":{},"pipes":[]}]}`
	}
	before := group(other)
	for _, mode := range []string{"system", "user"} {
		userField := ""
		if mode == "user" {
			userField = `,"user":"team$u"`
		}
		added := `{"id":" new ","source":{"bucket":"team/photos:marker","zones":["Zeta","Alpha"]},"dest":{"bucket":"*","zones":["*"]},"params":{"source":{"filter":{"tags":[]}},"dest":{},"priority":0,"mode":"` + mode + `"` + userField + `}}`
		after := group(other + "," + added)
		for _, tc := range []struct {
			name, initial, expected, actual, code string
			fail, count                           int
			zoneErr                               bool
		}{
			{"create", before, before, policy(after), "", 0, 3, false},
			{"empty pipes", group(""), group(""), policy(group(added)), "", 0, 3, false},
			{"duplicate ID", after, after, policy(after), "pre_check_failed", 0, 1, false},
			{"stale", before, group(""), policy(after), "pre_check_failed", 0, 1, false},
			{"malformed snapshot", before, "broken", policy(after), "pre_check_failed", 0, 1, false},
			{"pre read", before, before, policy(after), "pre_check_failed", 1, 1, false},
			{"mapping read", before, before, policy(after), "pre_check_failed", 0, 1, true},
			{"write", before, before, policy(after), "command_failed", 2, 2, false},
			{"post read", before, before, policy(after), "post_check_failed", 3, 3, false},
			{"unchanged", before, before, policy(before), "post_check_failed", 0, 3, false},
			{"old pipe changed", before, before, strings.Replace(policy(after), "9007199254740993", "9007199254740992", 1), "post_check_failed", 0, 3, false},
			{"other group removed", before, before, `{"groups":[` + after + `]}`, "post_check_failed", 0, 3, false},
			{"wrong default filter", before, before, strings.Replace(policy(after), `"tags":[]`, `"tags":[],"prefix":"unexpected"`, 1), "post_check_failed", 0, 3, false},
		} {
			for _, tenant := range []string{"", "team"} {
				t.Run(mode+"/"+tc.name+"/"+tenant, func(t *testing.T) {
					s, _, clusterID := newCephUserService(t)
					runner := &syncGroupExecutor{before: policy(tc.initial), after: tc.actual, failAt: tc.fail}
					wrapped := &syncFlowExecutor{syncGroupExecutor: runner}
					if tc.zoneErr {
						wrapped.zoneErr = errors.New("zonegroup unavailable")
					}
					s.executor = wrapped
					p := pipeCreateParameters()
					p["expected_group"] = tc.expected
					p["mode"] = mode
					p["bucket_id"] = base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos"))
					if mode == "user" {
						p["user"] = "team$u"
					}
					_, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_pipe_create", Parameters: p})
					if len(runner.calls) != tc.count {
						t.Fatalf("calls=%d expected=%d err=%v", len(runner.calls), tc.count, err)
					}
					if tc.count > 1 && wrapped.zoneCalls != 1 {
						t.Fatal("zone mapping omitted")
					}
					for i, call := range runner.calls {
						args := []string{"sync", "policy", "get"}
						if i == 1 {
							args = []string{"sync", "group", "pipe", "create", "--group-id", "g", "--pipe-id", " new ", "--source-zone-ids", "a,b", "--source-tenant", "team", "--source-bucket", "photos", "--source-bucket-id", "marker", "--dest-zone-ids", "*", "--dest-tenant", "*", "--dest-bucket", "*", "--dest-bucket-id", "*", "--mode", mode}
							if mode == "user" {
								args = append(args, "--uid", "team$u")
							}
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
	}
}

func TestBucketSyncPipeCreateValidation(t *testing.T) {
	for _, change := range []map[string]any{
		{"pipe_id": ""}, {"pipe_id": "-bad"}, {"expected_group": ""}, {"source_zones": []any{}}, {"source_zones": []any{"a", "a"}}, {"source_zones": []any{"*", "a"}}, {"dest_zones": []any{"a;b"}}, {"dest_zones": []any{"a=b"}}, {"source_bucket": ""}, {"source_bucket": "team/photos"}, {"source_bucket": "pho*"}, {"source_tenant": "team/x"}, {"dest_bucket_id": "m:id"}, {"mode": "unknown"}, {"mode": "user"}, {"user": "u"}, {"mode": "user", "user": "$u"}, {"mode": "user", "user": "team$"},
	} {
		p := pipeCreateParameters()
		for k, v := range change {
			p[k] = v
		}
		if _, err := bucketSyncPipeCreateArgs(p); err == nil {
			t.Fatalf("accepted invalid input: %+v", change)
		}
	}
	p := pipeCreateParameters()
	p["source_zones"] = []any{"*"}
	p["source_tenant"] = ""
	p["source_bucket"] = "*"
	delete(p, "source_bucket_id")
	if _, err := bucketSyncPipeCreateArgs(p); err != nil {
		t.Fatal(err)
	}
	g := map[string]any{"pipes": []any{}}
	if err := addBucketSyncPipe(g, p, []byte("{}")); err != nil {
		t.Fatal(err)
	}
	pipe := g["pipes"].([]any)[0].(map[string]any)
	if pipe["source"].(map[string]any)["bucket"] != "*" {
		t.Fatalf("wildcard bucket: %+v", pipe)
	}
	p = pipeCreateParameters()
	p["source_zones"] = []any{"unknown"}
	if err := addBucketSyncPipe(map[string]any{"pipes": []any{}}, p, []byte(`{"zones":[{"id":"a","name":"A"}]}`)); err == nil {
		t.Fatal("unknown zone accepted")
	}
}
