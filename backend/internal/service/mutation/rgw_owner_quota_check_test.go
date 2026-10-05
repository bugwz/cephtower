package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWOwnerQuotaReadback(t *testing.T) {
	for _, owner := range []struct{ action, field, id, identity, totalScope, totalField string }{
		{"rgw_user.quota", "uid", "tenant$namespace$user", "full_user_id", "user", "user_quota"},
		{"rgw_account.quota", "account_id", "RGW123", "id", "account", "quota"},
	} {
		for _, scope := range []string{owner.totalScope, "bucket"} {
			field := owner.totalField
			if scope == "bucket" {
				field = "bucket_quota"
			}
			for _, enabled := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%s/%t", owner.action, scope, enabled), func(t *testing.T) {
					params := map[string]any{owner.field: owner.id, "scope": scope, "enabled": enabled, "max_size": json.Number("1025"), "max_objects": json.Number("0")}
					validQuota := map[string]any{"enabled": enabled, "max_size": 2048, "max_objects": 0}
					for _, tc := range []struct {
						name   string
						mutate func(map[string]any)
						valid  bool
					}{
						{"matching", func(map[string]any) {}, true},
						{"wrong owner", func(v map[string]any) { v[owner.identity] = "other" }, false},
						{"missing owner", func(v map[string]any) { delete(v, owner.identity) }, false},
						{"wrong scope", func(v map[string]any) {
							v["user_quota"], v["quota"], v["bucket_quota"] = validQuota, validQuota, validQuota
							delete(v, field)
						}, false},
						{"unchanged size", func(v map[string]any) {
							v[field] = map[string]any{"enabled": enabled, "max_size": 1024, "max_objects": 0}
						}, false},
						{"wrong state", func(v map[string]any) {
							v[field] = map[string]any{"enabled": !enabled, "max_size": 2048, "max_objects": 0}
						}, false},
						{"missing objects", func(v map[string]any) { v[field] = map[string]any{"enabled": enabled, "max_size": 2048} }, false},
						{"missing state", func(v map[string]any) { v[field] = map[string]any{"max_size": 2048, "max_objects": 0} }, false},
						{"wrong objects", func(v map[string]any) {
							v[field] = map[string]any{"enabled": enabled, "max_size": 2048, "max_objects": 1}
						}, false},
					} {
						t.Run(tc.name, func(t *testing.T) {
							data := map[string]any{owner.identity: owner.id, field: validQuota}
							tc.mutate(data)
							raw, _ := json.Marshal(data)
							service, _, clusterID := newCephUserService(t)
							service.executor = bucketQuotaExecutor{raw}
							_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: owner.action, Parameters: params})
							if (err == nil) != tc.valid {
								t.Fatalf("valid=%t error=%v", tc.valid, err)
							}
							if !tc.valid {
								var actionErr *cephdomain.ActionError
								if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
									t.Fatalf("expected non-retryable post check failure: %v", err)
								}
							}
						})
					}
				})
			}
		}
	}
}

func TestRGWOwnerQuotaExactNumbers(t *testing.T) {
	for _, size := range []int64{-1, 0, 1, 1024, 9007199254740991} {
		rounded := size
		if size >= 0 {
			rounded = (size + 1023) / 1024 * 1024
		}
		params := map[string]any{"uid": "u", "scope": "user", "enabled": false, "max_size": json.Number(fmt.Sprint(size)), "max_objects": json.Number("9007199254740991")}
		raw := []byte(fmt.Sprintf(`{"full_user_id":"u","user_quota":{"enabled":false,"max_size":%d,"max_objects":9007199254740991}}`, rounded))
		if !rgwOwnerQuotaMatches("rgw_user.quota", params, raw) {
			t.Fatalf("rejected size %d", size)
		}
		for _, invalid := range []string{`null`, `{}`, `[]`, `not json`, `{"full_user_id":"u","user_quota":null}`, `{"full_user_id":"u","user_quota":{"enabled":false,"max_size":"0","max_objects":0}}`} {
			if rgwOwnerQuotaMatches("rgw_user.quota", params, []byte(invalid)) {
				t.Fatalf("accepted %s", invalid)
			}
		}
	}
}
