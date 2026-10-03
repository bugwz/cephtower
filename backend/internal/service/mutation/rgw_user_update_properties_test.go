package mutation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWUserUpdatePropertiesReadback(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	for _, tc := range []struct {
		field                string
		input, output, wrong any
	}{
		{"display_name", "New Name", "New Name", "Old"},
		{"email", "new@example.test", "new@example.test", "old@example.test"},
		{"email", "", "", "old@example.test"},
		{"max_buckets", json.Number("-1"), -1, 0},
		{"max_buckets", json.Number("0"), 0, -1},
		{"max_buckets", json.Number("2147483647"), 2147483647, 10},
		{"system", true, true, false}, {"system", false, false, true},
		{"suspended", true, 1, 0}, {"suspended", false, 0, 1},
	} {
		for _, scenario := range []string{"success", "missing", "null", "wrong", "wrong-user", "write-error", "read-error"} {
			info := map[string]any{"full_user_id": "tenant$ns$user", tc.field: tc.output}
			fail := ""
			count := 2
			switch scenario {
			case "missing":
				delete(info, tc.field)
			case "null":
				info[tc.field] = nil
			case "wrong":
				info[tc.field] = tc.wrong
			case "wrong-user":
				info["full_user_id"] = "other"
			case "write-error":
				fail = "rgw_user.update"
				count = 1
			case "read-error":
				fail = "rgw_user.update.post_check"
			}
			raw, _ := json.Marshal(info)
			runner := &directoryRenameExecutor{failID: fail, outputs: map[string]string{"rgw_user.update.post_check": string(raw)}}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/tenant$ns$user", Parameters: map[string]any{tc.field: tc.input}})
			if (err == nil) != (scenario == "success") || len(runner.specs) != count {
				t.Fatalf("%s=%v %s: %v", tc.field, tc.input, scenario, err)
			}
			if err != nil {
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Retryable || (count == 2 && failure.Code != "post_check_failed") {
					t.Fatalf("unsafe failure: %v", err)
				}
			}
		}
	}
}

func TestRGWUserUpdateCombinedProperties(t *testing.T) {
	service, _, clusterID := newCephUserService(t)
	p := map[string]any{"display_name": "Name", "email": "", "system": false, "suspended": true, "max_buckets": json.Number("0")}
	for _, fail := range []string{"", "rgw_user.update.step2"} {
		runner := &directoryRenameExecutor{failID: fail, outputs: map[string]string{"rgw_user.update.post_check": `{"full_user_id":"u","display_name":"Name","email":"","system":false,"suspended":1,"max_buckets":0}`}}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: clusterID, Action: "rgw_user.update", ResourceKey: "rgw/user/u", Parameters: p})
		if fail == "" {
			if err != nil || len(runner.specs) != 3 || runner.specs[1].Args[1] != "suspend" || runner.specs[2].Mutating {
				t.Fatalf("wrong combined result: %v %+v", err, runner.specs)
			}
		} else {
			var failure *cephdomain.ActionError
			if !errors.As(err, &failure) || failure.Retryable || len(runner.specs) != 2 {
				t.Fatalf("unsafe partial failure: %v", err)
			}
		}
	}
}
