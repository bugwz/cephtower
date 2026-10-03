package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWUserCreationProperties(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, limit := range []int64{-1, 0, 19, 2147483647} {
		for _, account := range []bool{false, true} {
			p := map[string]any{"uid": "tenant$ns$user", "display_name": "MyUser", "email": "new@example.test", "max_buckets": json.Number(strconv.FormatInt(limit, 10))}
			if account {
				p["account_id"], p["account_root"] = "RGW12345678901234567", false
			}
			for _, scenario := range []string{"success", "wrong-name", "missing-name", "null-name", "wrong-email", "missing-email", "null-email", "wrong-limit", "missing-limit", "null-limit", "fractional-limit", "string-limit"} {
				info := map[string]any{"full_user_id": "tenant$ns$user", "display_name": "MyUser", "email": "new@example.test", "max_buckets": limit, "keys": []any{}, "swift_keys": []any{}}
				if account {
					info["account_id"], info["type"] = p["account_id"], "rgw"
				}
				switch scenario {
				case "wrong-name":
					info["display_name"] = "Other"
				case "missing-name":
					delete(info, "display_name")
				case "null-name":
					info["display_name"] = nil
				case "wrong-email":
					info["email"] = "old@example.test"
				case "missing-email":
					delete(info, "email")
				case "null-email":
					info["email"] = nil
				case "wrong-limit":
					info["max_buckets"] = limit + 1
				case "missing-limit":
					delete(info, "max_buckets")
				case "null-limit":
					info["max_buckets"] = nil
				case "fractional-limit":
					info["max_buckets"] = 0.5
				case "string-limit":
					info["max_buckets"] = "19"
				}
				raw, _ := json.Marshal(info)
				runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_user.create.pre_check": "[]", "rgw_user.create.post_check": string(raw)}}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.create", ResourceKey: "rgw/user/tenant$ns$user", Parameters: p})
				if (err == nil) != (scenario == "success") || len(runner.specs) != 3 {
					t.Fatalf("account=%v limit=%d %s: %v", account, limit, scenario, err)
				}
				if err != nil {
					var failure *cephdomain.ActionError
					if !errors.As(err, &failure) || failure.Retryable || failure.Code != "post_check_failed" {
						t.Fatalf("unsafe failure: %v", err)
					}
				}
			}
		}
	}
}

func TestRGWUserCreationUnspecifiedDefaults(t *testing.T) {
	p := map[string]any{"uid": "u", "display_name": " User "}
	for _, raw := range []string{`{"full_user_id":"u","display_name":"User"}`, `{"full_user_id":"u","display_name":"User","max_buckets":1000,"email":""}`} {
		if !rgwUserCreatePropertiesMatch([]byte(raw), p) {
			t.Fatalf("inferred unspecified defaults: %s", raw)
		}
	}
	if rgwUserCreatePropertiesMatch([]byte(`{"full_user_id":"other","display_name":"User"}`), p) {
		t.Fatal("accepted wrong identity")
	}
}
