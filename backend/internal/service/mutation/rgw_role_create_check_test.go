package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRoleCreationVerifiesPropertiesAndDefaults(t *testing.T) {
	for _, custom := range []bool{false, true} {
		p := map[string]any{"name": "reader", "assume_role_policy": "{}"}
		row := map[string]any{"RoleName": "reader", "AccountId": "", "RoleId": "native-id", "AssumeRolePolicyDocument": "{}", "MaxSessionDuration": 3600, "Path": "/", "Description": ""}
		if custom {
			p["account_id"], row["AccountId"] = "RGW123", "RGW123"
			p["path"], row["Path"] = "/team/", "/team/"
			p["description"], row["Description"] = "read role", "read role"
			p["max_session_duration"], row["MaxSessionDuration"] = float64(7200), 7200
		}
		raw, _ := json.Marshal(row)
		if !rgwRoleCreateMatches(p, raw) {
			t.Fatal("valid role rejected")
		}
		for key, original := range row {
			for _, value := range []any{nil, false, []any{}} {
				row[key] = value
				bad, _ := json.Marshal(row)
				if rgwRoleCreateMatches(p, bad) {
					t.Fatalf("accepted invalid %s=%v", key, value)
				}
			}
			row[key] = original
		}
		for _, fail := range []string{"", "rgw_role.create", "rgw_role.create.post_check", "mismatch"} {
			s, _, id := newCephUserService(t)
			output := string(raw)
			if fail == "mismatch" {
				output = `{}`
			}
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_role.create.post_check": output}, failID: fail}
			s.executor = runner
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_role.create", Parameters: p})
			if (err == nil) != (fail == "") {
				t.Fatalf("%s: %v", fail, err)
			}
			if err != nil {
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Retryable {
					t.Fatalf("unsafe retry: %v", err)
				}
			}
			count := 2
			if fail == "rgw_role.create" {
				count = 1
			}
			if len(runner.specs) != count {
				t.Fatal("unexpected continuation")
			}
		}
		if !custom {
			if _, present := p["max_session_duration"]; present {
				t.Fatal("mutated request defaults")
			}
		}
	}
}
