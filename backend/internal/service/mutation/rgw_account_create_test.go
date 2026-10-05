package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWAccountCreateReadback(t *testing.T) {
	for _, tc := range []struct {
		raw, failID string
		valid       bool
	}{
		{`{"id":"RGW123","name":"","email":"","tenant":""}`, "", true},
		{`{"id":"RGW123","name":"","email":"","tenant":"","max_users":42}`, "", true},
		{`{"id":"other","name":"","email":"","tenant":""}`, "", false},
		{`{"id":"RGW123","name":"","email":"","tenant":"other"}`, "", false},
		{`{"id":"RGW123","name":"","email":""}`, "", false},
		{`{"id":"RGW123","name":null,"email":"","tenant":""}`, "", false},
		{`null`, "", false}, {`{}`, "", false}, {`invalid`, "", false},
		{`{}`, "rgw_account.create", false}, {`{}`, "rgw_account.create.post_check", false},
	} {
		service, _, id := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_account.create.post_check": tc.raw}, failID: tc.failID}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_account.create", Parameters: map[string]any{"account_id": "RGW123"}})
		if (err == nil) != tc.valid {
			t.Fatalf("valid=%t err=%v", tc.valid, err)
		}
		if !tc.valid {
			var actionErr *cephdomain.ActionError
			code := "post_check_failed"
			if tc.failID == "rgw_account.create" {
				code = "ceph_command_failed"
			}
			if !errors.As(err, &actionErr) || actionErr.Code != code || actionErr.Retryable {
				t.Fatalf("unsafe error: %v", err)
			}
		}
		count := 2
		if tc.failID == "rgw_account.create" {
			count = 1
		}
		if len(runner.specs) != count {
			t.Fatalf("commands=%v", runner.specs)
		}
		if count == 2 && (runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, []string{"account", "get", "--account-id", "RGW123", "--format", "json"})) {
			t.Fatalf("wrong read: %v", runner.specs)
		}
	}
}

func TestRGWAccountCreateExplicitFields(t *testing.T) {
	params := map[string]any{"account_id": "RGW123", "account_name": "new", "email": " mail ", "tenant": "team", "max_users": json.Number("-1"), "max_roles": json.Number("0"), "max_groups": json.Number("2147483647"), "max_buckets": json.Number("0"), "max_access_keys": json.Number("1")}
	actual := map[string]any{"id": "RGW123", "name": "new", "email": " mail ", "tenant": "team", "max_users": json.Number("-1"), "max_roles": json.Number("0"), "max_groups": json.Number("2147483647"), "max_buckets": json.Number("0"), "max_access_keys": json.Number("1")}
	raw, _ := json.Marshal(actual)
	if !rgwAccountCreateMatches(params, raw) {
		t.Fatal("rejected matching creation")
	}
	for key, value := range actual {
		delete(actual, key)
		raw, _ := json.Marshal(actual)
		if rgwAccountCreateMatches(params, raw) {
			t.Fatalf("accepted missing %s", key)
		}
		actual[key] = value
	}
	if params["email"] != " mail " || len(params) != 9 {
		t.Fatal("modified request")
	}
}
