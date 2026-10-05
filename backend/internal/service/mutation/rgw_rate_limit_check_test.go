package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWRateLimitReadback(t *testing.T) {
	for _, action := range []string{"rgw_user.ratelimit", "rgw_bucket.ratelimit"} {
		field, target := "user_ratelimit", []string{"--uid", "tenant$namespace$user", "--ratelimit-scope", "user"}
		if action == "rgw_bucket.ratelimit" {
			field, target = "bucket_ratelimit", []string{"--bucket", "photos", "--ratelimit-scope", "bucket", "--tenant", "team"}
		}
		for _, enabled := range []bool{true, false} {
			params := map[string]any{"uid": "tenant$namespace$user", "bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00photos")), "enabled": enabled, "max_read_ops": json.Number("0"), "max_write_ops": json.Number("1"), "max_read_bytes": json.Number("1024"), "max_write_bytes": json.Number("9007199254740991")}
			limits := map[string]any{"enabled": enabled, "max_read_ops": 0, "max_write_ops": 1, "max_read_bytes": 1024, "max_write_bytes": json.Number("9007199254740991")}
			raw, _ := json.Marshal(map[string]any{field: limits})
			if !rgwRateLimitMatches(action, params, raw) {
				t.Fatalf("rejected %s %t", action, enabled)
			}
			for _, key := range []string{"enabled", "max_read_ops", "max_write_ops", "max_read_bytes", "max_write_bytes"} {
				original := limits[key]
				for _, invalid := range []any{nil, "0", -1, 0.5, false, true, 19} {
					if key == "enabled" && invalid == enabled {
						continue
					}
					limits[key] = invalid
					bad, _ := json.Marshal(map[string]any{field: limits})
					if rgwRateLimitMatches(action, params, bad) {
						t.Fatalf("accepted %s=%v", key, invalid)
					}
				}
				delete(limits, key)
				bad, _ := json.Marshal(map[string]any{field: limits})
				if rgwRateLimitMatches(action, params, bad) {
					t.Fatalf("accepted missing %s", key)
				}
				limits[key] = original
			}
			wrongField := "bucket_ratelimit"
			if field == wrongField {
				wrongField = "user_ratelimit"
			}
			wrongScope, _ := json.Marshal(map[string]any{wrongField: limits})
			for _, data := range [][]byte{raw, wrongScope, []byte(`{}`), []byte(`null`), []byte(`not json`)} {
				service, _, id := newCephUserService(t)
				runner := &rateLimitExecutor{response: data}
				service.executor = runner
				_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: action, Parameters: params})
				if string(data) == string(raw) {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					var actionErr *cephdomain.ActionError
					if !errors.As(err, &actionErr) || actionErr.Code != "post_check_failed" || actionErr.Retryable {
						t.Fatalf("error=%v", err)
					}
				}
				if len(runner.specs) != 3 {
					t.Fatalf("commands=%v", runner.specs)
				}
				verb := "disable"
				if enabled {
					verb = "enable"
				}
				wantState := append(append([]string{"ratelimit", verb}, target...), "--format", "json")
				wantRead := append(append([]string{"ratelimit", "get"}, target...), "--format", "json")
				if !reflect.DeepEqual(runner.specs[1].Args, wantState) || !reflect.DeepEqual(runner.specs[2].Args, wantRead) || runner.specs[2].Mutating {
					t.Fatalf("wrong scoped commands: %v", runner.specs)
				}
			}
		}
	}
}
