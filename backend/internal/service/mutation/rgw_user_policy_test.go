package mutation

import (
	"reflect"
	"testing"
)

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
