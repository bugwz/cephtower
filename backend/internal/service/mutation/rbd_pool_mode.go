package mutation

import (
	"context"
	"encoding/json"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func (s *Service) executeRBDPoolMode(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	pool, mode := optional(request.Parameters, "pool"), optional(request.Parameters, "mode")
	if _, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true}); err != nil {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "pool mirroring mode outcome is uncertain; inspect the pool before any manual retry", Retryable: false}
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".post_check", Binary: executor.BinaryRBD, Args: []string{"mirror", "pool", "info", pool, "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	var info struct {
		Mode string `json:"mode"`
	}
	if err != nil || json.Unmarshal(result.Stdout, &info) != nil || info.Mode != mode {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "pool mirroring mode was not confirmed by readback; inspect the pool before any manual retry", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"pool": pool, "mode": mode, "verified": true}}, nil
}
