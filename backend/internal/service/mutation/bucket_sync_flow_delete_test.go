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

func TestBucketSyncFlowDeletion(t *testing.T) {
	for _, kind := range []string{"symmetrical", "directional"} {
		flow := `{"id":" target ","zones":["Zeta","Alpha"]}`
		fields := map[string]any{"flow_type": kind, "flow_id": " target "}
		extra := []string{"--flow-type", kind, "--flow-id", " target "}
		if kind == "directional" {
			flow = `{"source_zone":"Zeta","dest_zone":"Alpha"}`
			fields = map[string]any{"flow_type": kind, "source_zone": "a", "dest_zone": "b"}
			extra = []string{"--flow-type", kind, "--flow-id", "directional", "--source-zone-id", "a", "--dest-zone-id", "b"}
		}
		otherFlow := strings.ReplaceAll(strings.ReplaceAll(flow, " target ", "other"), "Zeta", "Gamma")
		group := func(data string) string {
			return `{"id":"g","status":"forbidden","data_flow":` + data + `,"pipes":[{"priority":9007199254740993}]}`
		}
		selected := group(`{"` + kind + `":[` + flow + `]}`)
		remaining := group(`{}`)
		multiple := group(`{"` + kind + `":[` + otherFlow + `,` + flow + `]}`)
		kept := group(`{"` + kind + `":[` + otherFlow + `]}`)
		duplicate := group(`{"` + kind + `":[` + flow + `,` + flow + `]}`)
		policy := func(g string) string {
			return `{"groups":[` + g + `,{"id":"other","status":"allowed","data_flow":{},"pipes":[]}]}`
		}
		for _, tc := range []struct {
			name, before, expected, after, code string
			fail, count                         int
			zoneErr                             bool
		}{
			{"last flow", selected, selected, policy(remaining), "", 0, 3, false},
			{"keep other flow", multiple, multiple, policy(kept), "", 0, 3, false},
			{"missing flow", remaining, remaining, policy(remaining), "pre_check_failed", 0, 1, false},
			{"ambiguous flow", duplicate, duplicate, policy(remaining), "pre_check_failed", 0, 1, false},
			{"stale", selected, remaining, policy(remaining), "pre_check_failed", 0, 1, false},
			{"malformed snapshot", selected, "broken", policy(remaining), "pre_check_failed", 0, 1, false},
			{"pre read", selected, selected, policy(remaining), "pre_check_failed", 1, 1, false},
			{"write", selected, selected, policy(remaining), "command_failed", 2, 2, false},
			{"post read", selected, selected, policy(remaining), "post_check_failed", 3, 3, false},
			{"unchanged", selected, selected, policy(selected), "post_check_failed", 0, 3, false},
			{"pipe changed", selected, selected, strings.Replace(policy(remaining), "9007199254740993", "9007199254740992", 1), "post_check_failed", 0, 3, false},
			{"other group removed", selected, selected, `{"groups":[` + remaining + `]}`, "post_check_failed", 0, 3, false},
			{"mapping unavailable", selected, selected, policy(remaining), "pre_check_failed", 0, 1, true},
		} {
			if tc.zoneErr && kind == "symmetrical" {
				continue
			}
			for _, tenant := range []string{"", "team"} {
				t.Run(kind+"/"+tc.name+"/"+tenant, func(t *testing.T) {
					s, _, clusterID := newCephUserService(t)
					runner := &syncGroupExecutor{before: policy(tc.before), after: tc.after, failAt: tc.fail}
					wrapped := &syncFlowExecutor{syncGroupExecutor: runner}
					if tc.zoneErr {
						wrapped.zoneErr = errors.New("mapping unavailable")
					}
					s.executor = wrapped
					p := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte(tenant + "\x00photos")), "group_id": "g", "expected_group": tc.expected}
					for k, v := range fields {
						p[k] = v
					}
					_, err := s.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_bucket.sync_flow_delete", Parameters: p})
					if len(runner.calls) != tc.count {
						t.Fatalf("calls=%d expected=%d err=%v", len(runner.calls), tc.count, err)
					}
					if kind == "symmetrical" && wrapped.zoneCalls != 0 {
						t.Fatal("symmetrical removal should not resolve zones")
					}
					if kind == "directional" && tc.count > 1 && wrapped.zoneCalls != 1 {
						t.Fatal("zone mapping not checked")
					}
					for i, call := range runner.calls {
						args := []string{"sync", "policy", "get"}
						if i == 1 {
							args = append([]string{"sync", "group", "flow", "remove", "--group-id", "g"}, extra...)
						}
						args = append(args, "--bucket", "photos", "--tenant", tenant, "--format", "json")
						if call.Binary != executor.BinaryRGWAdmin || call.Mutating != (i == 1) || !reflect.DeepEqual(args, call.Args) {
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
						t.Fatalf("error: %v", err)
					}
				})
			}
		}
	}
}

func TestBucketSyncFlowDeleteValidation(t *testing.T) {
	for _, p := range []map[string]any{
		{"flow_type": "symmetrical", "flow_id": "f"},
		{"flow_type": "symmetrical", "flow_id": "-bad", "expected_group": "{}"},
		{"flow_type": "symmetrical", "flow_id": "f", "expected_group": "{}", "zones": []any{"a"}},
		{"flow_type": "symmetrical", "flow_id": "f", "expected_group": "{}", "source_zone": "a"},
		{"flow_type": "unknown", "flow_id": "f", "expected_group": "{}"},
		{"flow_type": "directional", "source_zone": "a", "dest_zone": "a", "expected_group": "{}"},
	} {
		if _, err := bucketSyncFlowDeleteArgs(p); err == nil {
			t.Fatalf("accepted: %+v", p)
		}
	}
}
