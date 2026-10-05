package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

type bucketQuotaExecutor struct{ response []byte }

func (e bucketQuotaExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.Mutating {
		return executor.CommandResult{}, nil
	}
	return executor.CommandResult{Stdout: e.response}, nil
}

func TestBucketQuotaReadbackRejectsSilentWriteFailure(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		success        bool
	}{
		{"rounded", `{"bucket":"photos","tenant":"team","bucket_quota":{"enabled":true,"max_size":2048,"max_objects":0}}`, true},
		{"unchanged", `{"bucket":"photos","tenant":"team","bucket_quota":{"enabled":true,"max_size":1024,"max_objects":0}}`, false},
		{"wrong tenant", `{"bucket":"photos","tenant":"other","bucket_quota":{"enabled":true,"max_size":2048,"max_objects":0}}`, false},
		{"missing values", `{"bucket":"photos","tenant":"team","bucket_quota":{"enabled":true}}`, false},
		{"invalid json", `not json`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, _, id := newCephUserService(t)
			service.executor = bucketQuotaExecutor{[]byte(tc.response)}
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_bucket.quota", Parameters: map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00photos")), "enabled": true, "max_size": float64(1025), "max_objects": float64(0)}})
			if (err == nil) != tc.success {
				t.Fatalf("success=%v error=%v", tc.success, err)
			}
			if !tc.success {
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
					t.Fatalf("unsafe failure: %v", err)
				}
			}
		})
	}
}

func TestBucketQuotaDefaultTenantMustBeExplicit(t *testing.T) {
	params := map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("\x00photos")), "enabled": false, "max_size": json.Number("0"), "max_objects": json.Number("-1")}
	for _, tc := range []struct {
		identity string
		valid    bool
	}{
		{`"bucket":"photos","tenant":""`, true},
		{`"bucket":"photos"`, false},
		{`"bucket":"photos","tenant":null`, false},
		{`"bucket":"photos","tenant":false`, false},
		{`"bucket":"photos","tenant":"team"`, false},
		{`"bucket":null,"tenant":""`, false},
		{`"bucket":"other","tenant":""`, false},
	} {
		raw := `{` + tc.identity + `,"bucket_quota":{"enabled":false,"max_size":0,"max_objects":-1}}`
		if bucketQuotaMatches(params, []byte(raw)) != tc.valid {
			t.Fatalf("wrong match for %s", raw)
		}
	}
}

func TestBucketQuotaExecutionStopsOnUncertainState(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, failID := range []string{"", "rgw_bucket.quota", "rgw_bucket.quota.post_check", "rgw_bucket.quota.step2"} {
			if enabled && failID == "rgw_bucket.quota.step2" {
				continue
			}
			service, _, id := newCephUserService(t)
			raw, _ := json.Marshal(map[string]any{"bucket": "photos", "tenant": "team", "bucket_quota": map[string]any{"enabled": enabled, "max_size": 2048, "max_objects": 0}})
			runner := &directoryRenameExecutor{outputs: map[string]string{"rgw_bucket.quota.post_check": string(raw)}, failID: failID}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "rgw_bucket.quota", Parameters: map[string]any{"bucket_id": base64.RawURLEncoding.EncodeToString([]byte("team\x00photos")), "enabled": enabled, "max_size": json.Number("1025"), "max_objects": json.Number("0")}})
			if failID == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				code := "ceph_command_failed"
				if strings.HasSuffix(failID, ".post_check") {
					code = "post_check_failed"
				}
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Code != code || failure.Retryable || !strings.Contains(failure.Message, "inspect") {
					t.Fatalf("unsafe failure: %v", err)
				}
			}
			count := 3
			if enabled || failID == "rgw_bucket.quota.step2" {
				count = 2
			}
			if failID == "rgw_bucket.quota" {
				count = 1
			}
			if len(runner.specs) != count {
				t.Fatalf("unexpected continuation: %+v", runner.specs)
			}
			if failID == "" || strings.HasSuffix(failID, ".post_check") {
				read := runner.specs[count-1]
				if read.Mutating || !reflect.DeepEqual(read.Args, []string{"bucket", "stats", "--bucket", "photos", "--tenant", "team", "--format", "json"}) {
					t.Fatalf("wrong read scope: %+v", read)
				}
			}
		}
	}
}
