package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestRGWUserPolicyExecutionReadback(t *testing.T) {
	const arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
	service, _, clusterID := newCephUserService(t)
	for _, tc := range []struct {
		name, action, response, failID string
		wantSuccess                    bool
	}{
		{"attach verified", "attach", `["` + arn + `"]`, "", true},
		{"detach verified", "detach", `[]`, "", true},
		{"attach missing", "attach", `[]`, "", false},
		{"detach remains", "detach", `["` + arn + `"]`, "", false},
		{"other policy", "attach", `["arn:aws:iam::aws:policy/Other"]`, "", false},
		{"null list", "detach", `null`, "", false},
		{"duplicate target", "attach", `["` + arn + `","` + arn + `"]`, "", false},
		{"duplicate other", "detach", `["other","other"]`, "", false},
		{"wrong shape", "detach", `{"AttachedPolicies":[]}`, "", false},
		{"trailing data", "detach", `[] {}`, "", false},
		{"read failure", "attach", `["` + arn + `"]`, "rgw_user.policy.post_check", false},
		{"write failure", "attach", `["` + arn + `"]`, "rgw_user.policy", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.policy": "Managed policy operation succeeded", "rgw_user.policy.post_check": tc.response}, failID: tc.failID}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.policy", Parameters: map[string]any{"uid": "tenant$user", "action": tc.action, "policy_arn": arn}})
			if tc.wantSuccess {
				if err != nil {
					t.Fatal(err)
				}
			} else if tc.failID == "rgw_user.policy" {
				if err == nil {
					t.Fatal("write failure reported success")
				}
			} else {
				var actionErr *cephdomain.ActionError
				if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
					t.Fatalf("readback error = %v", err)
				}
			}
			wantCalls := 2
			if tc.failID == "rgw_user.policy" {
				wantCalls = 1
			}
			if len(runner.specs) != wantCalls {
				t.Fatalf("calls = %d, want %d", len(runner.specs), wantCalls)
			}
			write := runner.specs[0]
			if write.ID != "rgw_user.policy" || write.Binary != executor.BinaryRGWAdmin || !write.Mutating {
				t.Fatalf("write = %+v", write)
			}
			if wantCalls == 2 {
				read := runner.specs[1]
				if read.ID != "rgw_user.policy.post_check" || read.Binary != executor.BinaryRGWAdmin || read.Mutating || !reflect.DeepEqual(read.Args, []string{"user", "policy", "list", "attached", "--uid", "tenant$user", "--format", "json"}) {
					t.Fatalf("readback = %+v", read)
				}
			}
		})
	}
}

func TestRGWUserPolicyCommands(t *testing.T) {
	const arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
	if !Supports("rgw_user.policy") {
		t.Fatal("action unsupported")
	}
	for _, action := range []string{"attach", "detach"} {
		cmd, err := build(Request{Action: "rgw_user.policy"}, map[string]any{"uid": "tenant$user", "action": action, "policy_arn": arn})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cmd.args, []string{"user", "policy", action, "--uid", "tenant$user", "--policy-arn", arn, "--format", "json"}) {
			t.Fatalf("args = %v", cmd.args)
		}
		if !reflect.DeepEqual(cmd.check, []string{"user", "policy", "list", "attached", "--uid", "tenant$user", "--format", "json"}) {
			t.Fatalf("check = %v", cmd.check)
		}
	}
	for key, values := range map[string][]any{
		"uid":        {"", "--uid", "u\n", "u;other"},
		"action":     {"", "delete", "attach;detach", nil},
		"policy_arn": {"", "--policy", "arn:aws:iam::aws:policy/", arn + "\n", arn + ";other", nil},
	} {
		for _, value := range values {
			params := map[string]any{"uid": "tenant$user", "action": "attach", "policy_arn": arn}
			params[key] = value
			if _, err := build(Request{Action: "rgw_user.policy"}, params); err == nil {
				t.Fatalf("accepted %s = %#v", key, value)
			}
		}
	}
}

func TestRGWUserPolicyPostCheck(t *testing.T) {
	const arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
	for _, tc := range []struct {
		raw, action string
		want        bool
	}{
		{`["` + arn + `"]`, "attach", true},
		{`["` + arn + `"]`, "detach", false},
		{`[]`, "attach", false}, {`[]`, "detach", true},
		{`null`, "detach", false}, {`{}`, "detach", false},
		{`[null]`, "detach", false}, {`[1]`, "detach", false},
		{`[""]`, "detach", false}, {`invalid`, "attach", false},
	} {
		if got := rgwUserPolicyMatches([]byte(tc.raw), map[string]any{"policy_arn": arn, "action": tc.action}); got != tc.want {
			t.Fatalf("%s %s = %v", tc.action, tc.raw, got)
		}
	}
}
