package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWAccountUpdateFieldMatching(t *testing.T) {
	for _, field := range []string{"account_name", "email", "max_users", "max_roles", "max_groups", "max_buckets", "max_access_keys"} {
		key := field
		values := []any{json.Number("-1"), json.Number("0"), json.Number("2147483647")}
		if field == "account_name" || field == "email" {
			values = []any{"new value", " value with spaces "}
		}
		if field == "account_name" {
			key = "name"
		}
		for _, value := range values {
			params := map[string]any{"account_id": "RGW123", field: value}
			actual := map[string]any{"id": "RGW123", key: value}
			raw, _ := json.Marshal(actual)
			if !rgwAccountUpdateMatches(params, raw) {
				t.Fatalf("rejected %s=%v", field, value)
			}
			for _, invalid := range []any{nil, false, []any{}, map[string]any{}, "unexpected", json.Number("1.5")} {
				actual[key] = invalid
				raw, _ := json.Marshal(actual)
				if rgwAccountUpdateMatches(params, raw) {
					t.Fatalf("accepted invalid %s=%v", field, invalid)
				}
			}
			delete(actual, key)
			raw, _ = json.Marshal(actual)
			if rgwAccountUpdateMatches(params, raw) {
				t.Fatalf("accepted missing %s", field)
			}
		}
	}
}

func TestRGWAccountUpdateReadback(t *testing.T) {
	for _, tc := range []struct {
		raw, failID string
		valid       bool
	}{
		{`{"id":"RGW123","name":"new","max_buckets":0}`, "", true},
		{`{"id":"other","name":"new","max_buckets":0}`, "", false},
		{`{"name":"new","max_buckets":0}`, "", false},
		{`{"id":"RGW123","name":"old","max_buckets":0}`, "", false},
		{`{"id":"RGW123","name":"new","max_buckets":1}`, "", false},
		{`{"id":"RGW123","name":"new","max_buckets":"0"}`, "", false},
		{`null`, "", false}, {`[]`, "", false}, {`{}`, "", false}, {`invalid`, "", false},
		{`{}`, "rgw_account.update", false}, {`{}`, "rgw_account.update.post_check", false},
	} {
		service, _, clusterID := newCephUserService(t)
		runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_account.update.post_check": tc.raw}, failID: tc.failID}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_account.update", Parameters: map[string]any{"account_id": "RGW123", "account_name": "new", "max_buckets": json.Number("0")}})
		if (err == nil) != tc.valid {
			t.Fatalf("valid=%t err=%v", tc.valid, err)
		}
		if !tc.valid {
			var actionErr *cephdomain.ActionError
			code := "post_check_failed"
			if tc.failID == "rgw_account.update" {
				code = "ceph_command_failed"
			}
			if !errors.As(err, &actionErr) || actionErr.Code != code || actionErr.Retryable {
				t.Fatalf("unsafe error: %v", err)
			}
		}
		count := 2
		if tc.failID == "rgw_account.update" {
			count = 1
		}
		if len(runner.specs) != count {
			t.Fatalf("commands=%v", runner.specs)
		}
		if count == 2 && (runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[1].Args, []string{"account", "get", "--account-id", "RGW123", "--format", "json"})) {
			t.Fatalf("wrong read scope: %v", runner.specs)
		}
	}
}
