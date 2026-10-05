package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRoleUpdateVerifiesRequestedFields(t *testing.T) {
	for _, mode := range []string{"trust", "duration", "both"} {
		p := map[string]any{"name": "team$reader", "account_id": "RGW123"}
		if mode != "duration" {
			p["assume_role_policy"] = `{}`
		}
		if mode != "trust" {
			p["max_session_duration"] = float64(7200)
		}
		for _, field := range []string{"RoleName", "AccountId", "AssumeRolePolicyDocument", "MaxSessionDuration"} {
			for _, value := range []any{nil, false, "wrong", float64(3600)} {
				row := map[string]any{"RoleName": "team$reader", "AccountId": "RGW123", "AssumeRolePolicyDocument": "{}", "MaxSessionDuration": 7200}
				row[field] = value
				raw, _ := json.Marshal(row)
				// Unrequested fields may be absent or different, but malformed native types fail decoding.
				valid := (mode == "trust" && field == "MaxSessionDuration" && (value == nil || value == float64(3600))) || (mode == "duration" && field == "AssumeRolePolicyDocument" && (value == nil || value == "wrong"))
				if rgwRoleUpdateMatches(p, raw) != valid {
					t.Fatalf("%s %s %v", mode, field, value)
				}
			}
		}
		for _, fail := range []string{"", "rgw_role.update", "rgw_role.update.post_check", "rgw_role.update.step2"} {
			if fail == "rgw_role.update.step2" && mode != "both" {
				continue
			}
			s, _, id := newCephUserService(t)
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_role.update.post_check": `{"RoleName":"team$reader","AccountId":"RGW123","AssumeRolePolicyDocument":"{}","MaxSessionDuration":7200}`}, failID: fail}
			s.executor = runner
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_role.update", Parameters: p})
			if (err == nil) != (fail == "") {
				t.Fatalf("%s %s: %v", mode, fail, err)
			}
			if err != nil {
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Retryable {
					t.Fatalf("unsafe retry: %v", err)
				}
			}
			count := 2
			if mode == "both" {
				count = 3
			}
			if fail == "rgw_role.update" {
				count = 1
			}
			if fail == "rgw_role.update.step2" {
				count = 2
			}
			if len(runner.specs) != count {
				t.Fatal("unexpected continuation")
			}
		}
	}
}
