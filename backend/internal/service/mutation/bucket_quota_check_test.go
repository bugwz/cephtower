package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/base64"
	"testing"
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
		})
	}
}
